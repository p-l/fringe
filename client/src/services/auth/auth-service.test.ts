import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';
import AuthService, {useAuthService} from './auth-service';

describe('AuthService', () => {
  let mock : MockAdapter;

  beforeAll(() => {
    mock = new MockAdapter(axios);
    localStorage.removeItem('token');
    localStorage.removeItem('token_type');
    localStorage.removeItem('token_expires_at');
    localStorage.removeItem('token_role');
  });

  afterEach(() => {
    mock.reset();
  });

  it('calls back upon successful login', async () => {
    const auth = new AuthService();
    mock.onPost(auth.loginApiURL()).reply(200, {
      'token': 'a_token',
      'token_type': 'Bearer',
      'duration': 300,
    });

    auth.login('123', 'Bearer', (success, userAuth) =>{
      expect(success).toBe(true);
      expect(userAuth).not.toBeNull();
      expect(userAuth?.token).toBe('a_token');
      expect(userAuth?.tokenType).toBe('Bearer');
      expect(userAuth?.expires).toBeGreaterThan(Date.now());
    });
  });

  it('calls back with failure on auth failure', async () => {
    const auth = new AuthService();
    mock.onPost(auth.loginApiURL()).reply(400, {});

    auth.login('123', 'Bearer', (success, userAuth) =>{
      expect(success).toBe(false);
      expect(userAuth).toBeNull();
    });
  });

  it('calls back with failure on invalid response content', async () => {
    const auth = new AuthService();
    mock.onPost(auth.loginApiURL()).reply(200, {'token_type': 'Bearer'});

    auth.login('123', 'Bearer', (success, userAuth) =>{
      expect(success).toBe(false);
      expect(userAuth).toBeNull();
    });
  });

  it('calls back with failure on server error', async () => {
    const auth = new AuthService();
    mock.onPost(auth.loginApiURL()).reply(500);

    auth.login('123', 'Bearer', (success, userAuth) =>{
      expect(success).toBe(false);
      expect(userAuth).toBeNull();
    });
  });

  it('calls back on logout', async () => {
    const auth = new AuthService();

    const callback = jest.fn(()=>{});
    auth.logout(callback);

    expect(callback).toHaveBeenCalled();
    expect(mock.history.post.length).toBe(0);
    expect(mock.history.get.length).toBe(0);
    expect(mock.history.delete.length).toBe(0);
  });

  it('always return the same instance of AuthService', async () => {
    const oneAuth = useAuthService();
    const twoAuth = useAuthService();
    oneAuth.apiRootURL = '/test/api/';

    expect(Object.is(oneAuth, twoAuth)).toBe(true);
    expect(oneAuth.apiRootURL).toBe(twoAuth.apiRootURL);
  });

  it('does not double slash auth URL', async () => {
    const auth = new AuthService();
    expect(auth.apiRootURL).toContain('/api/');
    expect(auth.loginApiURL()).toContain('/api/auth/');

    auth.apiRootURL = '/test/api';
    expect(auth.apiRootURL).toContain('/test/api');
    expect(auth.loginApiURL()).toContain('/test/api/auth/');
  });

  it('stores and restores UserAuth', async () => {
    const auth = new AuthService();
    mock.onPost(auth.loginApiURL()).reply(200, {
      'token': 'a_token',
      'token_type': 'Bearer',
      'duration': 300,
    });

    auth.login('123', 'Bearer', (success, userAuth) =>{
      expect(success).toBe(true);
      expect(userAuth).not.toBeNull();
      expect(userAuth?.token).toBe('a_token');
      expect(userAuth?.tokenType).toBe('Bearer');
      expect(userAuth?.expires).toBeGreaterThan(Date.now());

      const currentAuth = auth.currentUserAuth;
      expect(currentAuth).not.toBeNull();
      expect(currentAuth?.token).toBe('a_token');
      expect(currentAuth?.tokenType).toBe('Bearer');
      expect(currentAuth?.expires).toBeGreaterThan(Date.now());
    });


    auth.logout(() => {});
    expect(auth.currentUserAuth).toBeNull();
  });


  it('does not return an expired auth after login', async () => {
    const auth = new AuthService();
    mock.onPost(auth.loginApiURL()).reply(200, {
      'token': 'a_token',
      'token_type': 'Bearer',
      'duration': -200,
    });

    auth.login('123', 'Bearer', (success, userAuth) =>{
      expect(success).toBe(true);
      expect(userAuth).not.toBeNull();
      expect(userAuth?.token).toBe('a_token');
      expect(userAuth?.tokenType).toBe('Bearer');
      expect(userAuth?.expires).toBeLessThan(Date.now()); // Token should be expired on reception

      const currentAuth = auth.currentUserAuth;
      expect(currentAuth).toBeNull(); // No token should be received since it'll never be stored
    });
  });

  it('does not return an expired auth after login', async () => {
    localStorage.setItem('token_type', 'test_type');
    localStorage.setItem('token', 'test_token');
    const futureExpiry = Date.now() + 300*1000;
    localStorage.setItem('token_expires_at', futureExpiry.toString());
    localStorage.setItem('token_role', 'user');

    const futureAuth = new AuthService();
    expect(futureAuth.currentUserAuth).not.toBeNull();

    const pastExpiry = Date.now() - 200*1000;
    localStorage.setItem('token_expires_at', pastExpiry.toString());

    const pastAuth = new AuthService();
    expect(pastAuth.currentUserAuth).toBeNull();
  });

  it('it adds authorization headers on calls to the API', async () => {
    localStorage.setItem('token_type', 'test_type');
    localStorage.setItem('token', 'test_token');
    const futureExpiry = Date.now() + 300 * 1000;
    localStorage.setItem('token_expires_at', futureExpiry.toString());

    const auth = useAuthService();
    mock.onPost(auth.loginApiURL()).reply(200, {
      'token': 'a_token',
      'token_type': 'Bearer',
      'duration': 300,
    });

    const targetURL = auth.apiRootURL + 'test/';
    mock.onGet(targetURL).reply((config) => {
      return [200, {requestHeaders: config.headers}];
    });

    auth.login('googleToken', 'googleTokenType', (success, userAuth) => {
      expect(success).toBeTruthy();
      expect(userAuth).not.toBeNull();

      axios.get(targetURL).then((response) => {
        expect(response).not.toBeNull();
        expect(response.data).not.toBeNull();
        expect(response.data.requestHeaders).not.toBeNull();
        expect(response.data.requestHeaders['Authorization']).not.toBeNull();
        expect(response.data.requestHeaders['Authorization']).toEqual(auth.currentUserAuth?.authorizationString());
      }).catch((error) => {
        console.error(error);
      });
    });
  });

  it('it adds authorization headers on calls to other sites', async () => {
    const targetURL = 'https://this.is.a.test.com/test/';
    mock.onGet(targetURL).reply((config) => {
      return [200, {requestHeaders: config.headers}];
    });

    axios.get(targetURL).then((response) => {
      expect(response).not.toBeNull();
      expect(response.data).not.toBeNull();
      expect(response.data.requestHeaders).not.toBeNull();
      expect(response.data.requestHeaders['Authorization']).toBeUndefined();
    }).catch((error) => {
      console.error(error);
    });
  });
});
