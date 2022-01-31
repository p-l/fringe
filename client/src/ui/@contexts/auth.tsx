import React from 'react';

interface AuthContextType {
  authenticated: boolean;
  login: (tokenType: string, token: string, callback: (authenticated: boolean) => void) => void;
  logout: (callback: VoidFunction) => void;
}

export function useAuth() {
  return React.useContext(AuthContext);
}

export const AuthContext = React.createContext<AuthContextType>(null!);
