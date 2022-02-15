import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';
import {useUserService, UserService} from './user-service';

describe('UserService', () => {
  let mock : MockAdapter;

  beforeAll(() => {
    mock = new MockAdapter(axios);
  });

  afterEach(() => {
    mock.reset();
  });

  it('always return the same instance of UserService', async () => {
    const oneUserService = useUserService();
    const twoUserService = useUserService();
    oneUserService.apiRootURL = '/test/api/';

    expect(Object.is(oneUserService, twoUserService)).toBe(true);
    expect(oneUserService.apiRootURL).toBe(twoUserService.apiRootURL);
  });

  it('does not double slash auth URL', async () => {
    const userService = useUserService();
    expect(userService.apiRootURL).toContain('/api/');
    expect(userService.userApiURL()).toContain('/api/users/');

    userService.apiRootURL = '/test/api';
    expect(userService.apiRootURL).toContain('/test/api');
    expect(userService.userApiURL()).toContain('/test/api/users/');
  });

  it('returns null on forbidden error', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/').reply(403, null, null);

    userService.me((user) => {
      expect(user).toBeNull();
    });
  });


  it('returns null on internal server error', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/').reply(500, null, null);

    userService.me((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns null when email field is missing', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/').reply(200, {
      'name': 'some user',
      'picture': 'https://picture.url',
      'password': 'hey look a password!',
      'password_updated_at': 0,
      'last_seen_at': 0,
    }, null); // missing email field

    userService.me((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns user when minimal fields are present', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/').reply(200, {
      'email': 'some@email.com',
      'name': 'some user',
      'picture': 'https://picture.url',
      'password_updated_at': 0,
      'last_seen_at': 0,
    }, null);

    userService.me((user) => {
      expect(user).not.toBeNull();
      expect(user?.email).toBe('some@email.com');
      expect(user?.lastSeenAt).toEqual(new Date(0));
    });
  });

  it('returns user password when received on initial get', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/').reply(200, {
      'email': 'some@email.com',
      'name': 'some user',
      'picture': 'https://picture.url',
      'password_updated_at': 0,
      'last_seen_at': 0,
      'password': 'super_secret',
    }, null);

    userService.me((user) => {
      expect(user).not.toBeNull();
      expect(user?.password).toEqual('super_secret');
    });
  });

  it('returns null when error is received on renew', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/renew/').reply(403, null, null); // missing email field

    userService.renewMyPassword((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns null when receiving invalid on renew', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/renew/').reply(200, {
      'name': 'some user',
      'picture': 'https://picture.url',
      'password_updated_at': 0,
      'last_seen_at': 0,
    }, null); // missing email field

    userService.renewMyPassword((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns null when empty password is received on renew', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/renew/').reply(200, {
      'email': 'some@email.com',
      'name': 'some user',
      'picture': 'https://picture.url',
      'password_updated_at': 0,
      'last_seen_at': 0,
      'password': '',
    }, null);

    userService.renewMyPassword((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns null when no password is received on renew', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/renew/').reply(200, {
      'email': 'some@email.com',
      'name': 'some user',
      'picture': 'https://picture.url',
      'password_updated_at': 0,
      'last_seen_at': 0,
    }, null);

    userService.renewMyPassword((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns user when a password is received on renew', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/renew/').reply(200, {
      'email': 'some@email.com',
      'name': 'some user',
      'picture': 'https://picture.url',
      'password_updated_at': 0,
      'last_seen_at': 0,
      'password': 'super_random_password',
    }, null);

    userService.renewMyPassword((user) => {
      expect(user).not.toBeNull();
      expect(user?.password).toEqual('super_random_password');
    });
  });

  it('returns a list of user model', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()).reply(200, [{
      'email': 'some@email.com',
      'name': 'some user',
      'picture': 'https://picture.url',
      'password_updated_at': 0,
      'last_seen_at': 0,
      'password': 'super_random_password',
    }], null);

    userService.findAllUsers('', 0, 20, (users, success) => {
      expect(success).toBeTruthy();
      expect(users).not.toBeNull();
      expect(users.length).toBe(1);
      expect(users[0].email).toEqual('some@email.com');
    });
  });

  it('catches error and return failure', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()).reply(500, null, null); // missing email field

    userService.findAllUsers('', 0, 20, (users, success) => {
      expect(success).not.toBeTruthy();
      expect(users).not.toBeNull();
      expect(users.length).toBe(0);
    });
  });

  it('returns a valid user when creating', async () => {
    const userService = new UserService();
    mock.onPost(userService.userApiURL()).reply(200, {
      result: 'success',
      user: {
        'email': 'some@email.com',
        'name': 'some user',
        'picture': 'https://picture.url',
        'password_updated_at': 0,
        'last_seen_at': 0,
        'password': 'super_random_password',
      }}, null);

    userService.create('some@email.com', 'some user', (resultText, user) => {
      expect(resultText).toEqual('success');
      expect(user).not.toBeNull();
      expect(user?.email).toEqual('some@email.com');
    });
  });

  it('returns a failure when receiving an invalid user', async () => {
    const userService = new UserService();
    mock.onPost(userService.userApiURL()).reply(200, {
      result: 'success',
      user: {
        'name': 'some user',
        'picture': 'https://picture.url',
        'password_updated_at': 0,
        'last_seen_at': 0,
        'password': 'super_random_password',
      }}, null); // no email for user

    userService.create('some@email.com', 'some user', (resultText, user) => {
      expect(resultText).toEqual('failed');
      expect(user).toBeNull();
    });
  });

  it('returns a failure when service returns an failure', async () => {
    const userService = new UserService();
    mock.onPost(userService.userApiURL()).reply(200, {
      result: 'exists',
      user: null,
    }, null);

    userService.create('some@email.com', 'some user', (resultText, user) => {
      expect(resultText).toEqual('exists');
      expect(user).toBeNull();
    });
  });

  it('returns a failure when service returns an error ', async () => {
    const userService = new UserService();
    mock.onPost(userService.userApiURL()).reply(500, null, null);

    userService.create('some@email.com', 'some user', (resultText, user) => {
      expect(resultText).toEqual('failed');
      expect(user).toBeNull();
    });
  });

  it('return success with no user when delete succeed', async () => {
    const userService = new UserService();
    mock.onDelete(userService.userApiURL()+'some%40email.com/').reply(200, {
      result: 'success',
      user: null,
    }, null);

    userService.delete('some@email.com', (resultText) => {
      expect(resultText).toEqual('success');
    });
  });

  it('return failed on invalid response data', async () => {
    const userService = new UserService();
    mock.onDelete(userService.userApiURL()+'some%40email.com/').reply(200, {
      user: null,
    }, null);

    userService.delete('some@email.com', (resultText) => {
      expect(resultText).toEqual('failed');
    });
  });

  it('return failed on error code', async () => {
    const userService = new UserService();
    mock.onDelete(userService.userApiURL()+'some%40email.com/').reply(500, null, null);

    userService.delete('some@email.com', (resultText) => {
      expect(resultText).toEqual('failed');
    });
  });
});
