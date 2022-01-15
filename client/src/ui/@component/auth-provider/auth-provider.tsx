import React from 'react';
import {useAuthService} from '../../../services/auth';
import {AuthContext} from '../../@contexts/auth';
import {UserAuth} from '../../../types/user-auth';

function AuthProvider({children}: { children: JSX.Element }) {
  const [userAuth, setUserAuth] = React.useState<UserAuth|null>(null);
  const authService = useAuthService();

  const login = (token: string, tokenType: string, callback: (success: boolean, auth: UserAuth|null) => void) => {
    return authService.login(token, tokenType, (success, auth) => {
      console.log(`Login attempt to Fringe Server: ${success ? 'success!' : 'failed'}`);
      if (success) {
        setUserAuth(auth);
        console.log(`token: ${auth?.tokenType} ${auth?.token}`);
      }
      callback(success, auth);
    });
  };

  const logout = (callback: VoidFunction) => {
    return authService.logout(() => {
      console.log('AuthProvider: logout!');
      setUserAuth(null);
      callback();
    });
  };

  const value = {userAuth, login, logout};

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export default AuthProvider;

