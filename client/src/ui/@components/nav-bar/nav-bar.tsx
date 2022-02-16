import {GroupRounded, LogoutRounded} from '@mui/icons-material';
import {AppBar, Box, Button, Stack, Toolbar} from '@mui/material';
import React from 'react';
import {GoogleLogout} from 'react-google-login';
import {useNavigate} from 'react-router-dom';

import Config from '../../../services/config';
import {useAuth} from '../../@contexts/auth';


function NavBar({config}: {config: Config}) {
  const auth = useAuth();
  const navigate = useNavigate();

  const googleLogoutSuccessHandler = () => {
    auth.logout(()=>{
      navigate('/login', {replace: true});
    });
  };

  const goToRoot = () => {
    navigate('/', {replace: true});
  };

  const goToAdmin = () => {
    navigate('/admin', {replace: true});
  };

  if (auth && auth.authenticated) {
    return (
      <AppBar position="sticky">
        <Toolbar disableGutters>
          <Box sx={{mr: 2}}>
            <Button onClick={goToRoot}>
              <img alt="fringe logo" src="/logo/fringe-user-white.svg" width={28}/>
            </Button>
          </Box>
          <Box sx={{flexGrow: 1}} />
          <Stack direction="row" justifyContent="flex-end" alignItems="center" spacing={0}>
            {auth.adminRole && (
              <Button color="inherit" onClick={goToAdmin}>
                <GroupRounded />
              </Button>
            )}
            <GoogleLogout
              clientId={config.googleClientID}
              onLogoutSuccess={googleLogoutSuccessHandler}
              render={(props: {onClick: () => void, disabled?: boolean | undefined}) => (
                <Button color="inherit" onClick={props.onClick} disabled={props.disabled}>
                  <LogoutRounded />
                </Button>
              )}/>
          </Stack>
        </Toolbar>
      </AppBar>
    );
  }

  return (<AppBar />);
}

export default NavBar;


