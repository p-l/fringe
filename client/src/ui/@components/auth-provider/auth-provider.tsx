import React from 'react';
import {useAuthService} from '../../../services/auth';
import {AuthContext} from '../../@contexts/auth';

function AuthProvider({children}: { children: JSX.Element[] }) {
  const authService = useAuthService();
  const currentUserAuth = authService.currentUserAuth;
  const [authenticated, setAuthenticated] = React.useState<boolean>(currentUserAuth != null);


  const login = (token: string, tokenType: string, callback: (authenticated: boolean) => void) => {
    return authService.login(token, tokenType, (success, _auth) => {
      setAuthenticated(success);
      callback(success);
    });
  };

  const logout = (callback: VoidFunction) => {
    return authService.logout(() => {
      setAuthenticated(false);
      callback();
    });
  };

  const value = {authenticated, login, logout};

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export default AuthProvider;

