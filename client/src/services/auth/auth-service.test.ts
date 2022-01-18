import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';
import AuthService, {useAuthService} from './auth-service';

describe('AuthService', () => {
  let mock : MockAdapter;

  beforeAll(() => {
    mock = new MockAdapter(axios);
  });

  afterEach(() => {
    mock.reset();
  });

  it('calls back upon successful login', async () => {
    const auth = new AuthService();
    mock.onPost(auth.loginApiURL()).reply(200, {
      'token': 'a_token',
      'token_type': 'Bearer',
    });

    auth.login('123', 'Bearer', (success, userAuth) =>{
      expect(success).toBe(true);
      expect(userAuth).not.toBeNull();
      expect(userAuth?.token).toBe('a_token');
      expect(userAuth?.tokenType).toBe('Bearer');
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
    expect(auth.apiRootURL).toBe('/api/');
    expect(auth.loginApiURL()).toBe('/api/auth/');

    auth.apiRootURL = '/test/api';
    expect(auth.apiRootURL).toBe('/test/api');
    expect(auth.loginApiURL()).toBe('/test/api/auth/');
  });
});
