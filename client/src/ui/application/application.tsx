import React, {Suspense} from 'react';
import {Route, Routes} from 'react-router-dom';
import {Box, CircularProgress, Container} from '@mui/material';
import {useUserService} from '../../services/user/user-service';

import ErrorBoundary from '../@components/error-boundary';
import AuthProvider from '../@components/auth-provider';
import NavBar from '../@components/nav-bar';
import RequireAuth from '../@components/require-auth';
import useMountEffect from '../@hooks/use-mount';
import Me from '../me';
import Login from '../login';
import Admin from '../admin';
import Config from '../../services/config';
import {useAuthService} from '../../services/auth';

interface ApplicationProps {
    config: Config
}

function Application(props: ApplicationProps) {
  useMountEffect(()=>{
    const authService = useAuthService();
    const userService = useUserService();
    authService.apiRootURL = props.config.apiRootURL;
    userService.apiRootURL = props.config.apiRootURL;
  });
  return (
    <React.StrictMode>
      <ErrorBoundary>
        <Suspense fallback={
          <Container maxWidth="xs">
            <Box sx={{p: 50, m: 50, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
              <CircularProgress />
            </Box>
          </Container>
        }>
          <AuthProvider>
            <NavBar config={props.config}/>
            <Routes>
              <Route path="/" element={<RequireAuth><Me /></RequireAuth>} />
              <Route path="/admin" element={<RequireAuth><Admin /></RequireAuth>} />
              <Route path="/login" element={<Login googleClientID={props.config.googleClientID} />} />
            </Routes>
          </AuthProvider>
        </Suspense>
      </ErrorBoundary>
    </React.StrictMode>
  );
}

export default Application;
