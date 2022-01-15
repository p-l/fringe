import React from 'react';
import {UserAuth} from '../../types/user-auth';

interface AuthContextType {
  userAuth: UserAuth|null;
  login: (token: string, tokenType: string, callback: (success: boolean, auth: UserAuth|null) => void) => void;
  logout: (callback: VoidFunction) => void;
}

export function useAuth() {
  return React.useContext(AuthContext);
}

export const AuthContext = React.createContext<AuthContextType>(null!);
