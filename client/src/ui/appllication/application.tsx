import React, {Suspense} from 'react';
import {Route, Routes} from 'react-router-dom';
import {Box, CircularProgress, Container} from '@mui/material';

import ErrorBoundary from '../@component/error-boundary';
import AuthProvider from '../@component/auth-provider';
import RequireAuth from '../@component/require-auth';
import {useAuthService} from '../../services/auth';
import Home from '../home';
import Login from '../login';
import Config from '../../config';

interface ApplicationProps {
    config: Config
}

function Application(props: ApplicationProps) {
  const authService = useAuthService();
  authService.authApiRootURL = `${props.config.apiURL}/auth/`;

  return (
    <ErrorBoundary>
      <Suspense fallback={
        <Container component="main" maxWidth="xs">
          <Box sx={{p: 50, m: 50, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
            <CircularProgress />
          </Box>
        </Container>
      }>
        <AuthProvider>
          <Routes>
            <Route path="/" element={<RequireAuth><Home config={props.config} /></RequireAuth>} />
            <Route path="/login" element={<Login googleClientID={props.config.googleClientID} />} />
          </Routes>
        </AuthProvider>
      </Suspense>
    </ErrorBoundary>
  );
}

export default Application;
