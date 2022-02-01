import {UserAuth} from './user-auth';

describe('UserAuth', () => {
  it('return properly formed authorization string', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now()+1000);
    expect(auth.token).toEqual('token');
    expect(auth.tokenType).toEqual('token-type');
    expect(auth.authorizationString()).toEqual('token-type token');
  });

  it('return expired on an expired token', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now());
    expect(auth.isExpired()).toBeTruthy();
  });

  it('return not expired on an unexpired token', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now()+300*1000);
    expect(auth.isExpired()).not.toBeTruthy();
  });
});
