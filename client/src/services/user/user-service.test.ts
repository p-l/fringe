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
      'email': 'email@email.com',
      'last_seen_at': 0,
    }, null); // missing email field

    userService.me((user) => {
      expect(user).not.toBeNull();
      expect(user?.email).toBe('email@email.com');
      expect(user?.lastSeenAt).toEqual(new Date(0));
    });
  });

  it('returns user password when received on initial get', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/').reply(200, {
      'email': 'email@email.com',
      'last_seen_at': 0,
      'password': 'super_secret',
    }, null); // missing email field

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
      'last_seen_at': 0,
      'password': '',
    }, null); // missing email field

    userService.renewMyPassword((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns null when empty password is received on renew', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/renew/').reply(200, {
      'email': 'email@email.com',
      'last_seen_at': 0,
      'password': '',
    }, null); // missing email field

    userService.renewMyPassword((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns null when no password is received on renew', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/renew/').reply(200, {
      'email': 'email@email.com',
      'last_seen_at': 0,
    }, null); // missing email field

    userService.renewMyPassword((user) => {
      expect(user).toBeNull();
    });
  });

  it('returns user when a password is received on renew', async () => {
    const userService = new UserService();
    mock.onGet(userService.userApiURL()+'me/renew/').reply(200, {
      'email': 'email@email.com',
      'last_seen_at': 0,
      'password': 'super_random_password',
    }, null); // missing email field

    userService.renewMyPassword((user) => {
      expect(user).not.toBeNull();
      expect(user?.password).toEqual('super_random_password');
    });
  });
});
