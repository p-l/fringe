import React from 'react';
import {Navigate} from 'react-router-dom';
import {useAuth} from '../../@contexts/auth';

function RequireAuth({children}: {children: JSX.Element}) {
  const auth = useAuth();
  console.log('RequireAuth!');

  if (!auth.userAuth) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

export default RequireAuth;

