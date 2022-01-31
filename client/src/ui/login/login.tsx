import React from 'react';
import {useNavigate} from 'react-router-dom';
import {Box, Button, Container, Paper, Typography} from '@mui/material';
import GoogleIcon from '@mui/icons-material/Google';
import GoogleLogin, {GoogleLoginResponse, GoogleLoginResponseOffline} from 'react-google-login';
import {useAuth} from '../@contexts/auth';

// const Login: React.FC<{googleClientID: string}> = ({children, googleClientID}) => {
function Login({googleClientID} : {googleClientID: string}) {
  const navigate = useNavigate();
  const auth = useAuth();

  const googleSuccessHandler = (response: (GoogleLoginResponse | GoogleLoginResponseOffline)) => {
    console.log(response);
    if ('accessToken' in response) {
      const token = response.accessToken;
      const tokenType = 'Bearer';

      auth.login(token, tokenType, (authenticate)=>{
        if (authenticate) {
          console.debug('Authentication was a success. Navigating to /');
          navigate('/', {replace: true});
        } else {
          // TODO: Display error banner
          throw new Error('Authentication rejected by Fringe server.');
        }
      });
    } else {
      googleFailureHandler(Error('invalid response from Google'));
    }
  };

  const googleFailureHandler = (error: any) => {
    // TODO: Display error banner
    console.log(error);
  };

  return (
    <Container component="main" maxWidth="xs">
      <Paper sx={{marginTop: 10, p: 2, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
        <Box sx={{m: 2, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
          <img alt="fringe logo" src="/logo/logo-blue.svg" width={128}/>
        </Box>
        <Box sx={{marginBottom: 2}}>
          <Typography component="h1" variant="h5">
            Welcome to Fringe
          </Typography>
        </Box>
        <Box sx={{m: 2}}>
          <GoogleLogin
            clientId={googleClientID}
            onSuccess={googleSuccessHandler}
            onFailure={googleFailureHandler}
            render={(props: {onClick: () => void, disabled?: boolean | undefined}) => (
              <Button size="large" variant="contained" startIcon={<GoogleIcon />} onClick={props.onClick} disabled={props.disabled}>
                <Typography>Sign-in With Google</Typography>
              </Button>
            )} />
        </Box>
      </Paper>
    </Container>
  );
}

export default Login;
