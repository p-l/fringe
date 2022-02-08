import {UserAuth, UserAuthRole} from './user-auth';

describe('UserAuth', () => {
  it('return properly formed authorization string', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now()+1000, 'user');
    expect(auth.token).toEqual('token');
    expect(auth.tokenType).toEqual('token-type');
    expect(auth.authorizationString()).toEqual('token-type token');
    expect(auth.roleString).toEqual('user');
    expect(auth.role).toEqual(UserAuthRole.user);
  });

  it('return expired on an expired token', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now(), 'admin');
    expect(auth.isExpired()).toBeTruthy();
  });

  it('return not expired on an unexpired token', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now()+300*1000, 'admin');
    expect(auth.isExpired()).not.toBeTruthy();
  });

  it('maps admin role to enum', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now()+300*1000, 'admin');
    expect(auth.role).toEqual(UserAuthRole.admin);
    expect(auth.roleString).toEqual('admin');
  });

  it('maps user role to enum', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now()+300*1000, 'user');
    expect(auth.role).toEqual(UserAuthRole.user);
    expect(auth.roleString).toEqual('user');
  });

  it('maps unknown role to enum', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now()+300*1000, 'something');
    expect(auth.role).toEqual(UserAuthRole.unknown);
    expect(auth.roleString).toEqual('unknown');
  });

  it('setting role enum maps to string', async () => {
    const auth = new UserAuth('token-type', 'token', Date.now()+300*1000, 'unknown');
    expect(auth.role).toEqual(UserAuthRole.unknown);
    expect(auth.roleString).toEqual('unknown');

    auth.role = UserAuthRole.user;
    expect(auth.role).toEqual(UserAuthRole.user);
    expect(auth.roleString).toEqual('user');

    auth.role = UserAuthRole.admin;
    expect(auth.role).toEqual(UserAuthRole.admin);
    expect(auth.roleString).toEqual('admin');

    auth.role = UserAuthRole.unknown;
    expect(auth.role).toEqual(UserAuthRole.unknown);
    expect(auth.roleString).toEqual('unknown');
  });
});
