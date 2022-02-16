import React from 'react';
import {UserAuth, UserAuthRole} from '../../../models/user-auth';
import {useAuthService} from '../../../services/auth';
import {AuthContext} from '../../@contexts/auth';

function AuthProvider({children}: { children: JSX.Element[] }) {
  const authService = useAuthService();
  const currentUserAuth = authService.currentUserAuth;
  const [authenticated, setAuthenticated] = React.useState<boolean>(currentUserAuth != null);
  const [adminRole, setAdminRole] = React.useState<boolean>(currentUserAuth?.role == UserAuthRole.admin);


  const login = (token: string, tokenType: string, callback: (authenticated: boolean) => void) => {
    return authService.login(token, tokenType, (success: boolean, auth: UserAuth|null) => {
      setAuthenticated(success);
      if (auth != null && auth.role == UserAuthRole.admin) {
        setAdminRole(true);
      } else {
        setAdminRole(false);
      }

      callback(success);
    });
  };

  const logout = (callback: VoidFunction) => {
    return authService.logout(() => {
      setAuthenticated(false);
      callback();
    });
  };

  const value = {authenticated, adminRole, login, logout};

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export default AuthProvider;

