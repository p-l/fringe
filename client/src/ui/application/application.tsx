import React, {Suspense} from 'react';
import {Route, Routes} from 'react-router-dom';
import {Box, CircularProgress, Container} from '@mui/material';

import ErrorBoundary from '../@components/error-boundary';
import AuthProvider from '../@components/auth-provider';
import NavBar from '../@components/nav-bar';
import RequireAuth from '../@components/require-auth';
import useMountEffect from '../@hooks/use-mount';
import Home from '../home';
import Login from '../login';
import Config from '../../services/config';
import {useAuthService} from '../../services/auth';

interface ApplicationProps {
    config: Config
}

function Application(props: ApplicationProps) {
  useMountEffect(()=>{
    const authService = useAuthService();
    authService.apiRootURL = props.config.apiRootURL;
  });


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
          <NavBar config={props.config}/>
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
