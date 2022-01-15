import React from 'react';
import {Navigate, useLocation} from 'react-router-dom';
import {useAuth} from '../../@contexts/auth';

function RequireAuth({children}: {children: JSX.Element}) {
  const auth = useAuth();
  console.log('RequireAuth!');
  const location = useLocation();

  if (!auth.userAuth) {
    return <Navigate to="/login" state={{from: location}} replace />;
  }

  return children;
}

export default RequireAuth;

