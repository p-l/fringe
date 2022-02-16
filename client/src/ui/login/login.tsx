import React from 'react';
import {Alert, Box, Button, Container, Paper, Snackbar, Typography} from '@mui/material';
import GoogleIcon from '@mui/icons-material/Google';
import {useNavigate} from 'react-router-dom';
import GoogleLogin, {GoogleLoginResponse, GoogleLoginResponseOffline} from 'react-google-login';
import {Trans, useTranslation} from 'react-i18next';

import {useAuth} from '../@contexts/auth';

// const Login: React.FC<{googleClientID: string}> = ({children, googleClientID}) => {
function Login({googleClientID} : {googleClientID: string}) {
  const [errorMessage, setErrorMessage] = React.useState<string>('');
  const navigate = useNavigate();
  const auth = useAuth();
  const {t} = useTranslation();

  const googleSuccessHandler = (response: (GoogleLoginResponse | GoogleLoginResponseOffline)) => {
    setErrorMessage('');
    console.log(response);
    if ('accessToken' in response) {
      const token = response.accessToken;
      const tokenType = 'Bearer';

      auth.login(token, tokenType, (authenticate)=>{
        if (authenticate) {
          console.debug('Authentication was a success. Navigating to /');
          navigate('/', {replace: true});
        } else {
          setErrorMessage(t('login.errorFringeRejected'));
        }
      });
    } else {
      googleFailureHandler(Error('Missing accessToken in Google\'s response'));
    }
  };

  const googleFailureHandler = (error: any) => {
    console.debug(error);
    setErrorMessage(t('admin.errorGoogleFailed'));
  };

  return (
    <Container component="main" maxWidth="xs">
      <Paper sx={{marginTop: 10, p: 2, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
        <Box sx={{m: 2, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
          <img alt="fringe logo" src="/logo/logo-blue.svg" width={128}/>
        </Box>
        <Box sx={{marginBottom: 2}}>
          <Typography component="h1" variant="h5">
            <Trans i18nKey='login.welcome' />
          </Typography>
        </Box>
        <Box sx={{m: 2}}>
          <GoogleLogin
            clientId={googleClientID}
            onSuccess={googleSuccessHandler}
            onFailure={googleFailureHandler}
            render={(props: {onClick: () => void, disabled?: boolean | undefined}) => (
              <Button size="large" variant="contained" startIcon={<GoogleIcon />} onClick={props.onClick} disabled={props.disabled}>
                <Typography><Trans i18nKey='login.signInButton' /></Typography>
              </Button>
            )} />
        </Box>
        <Snackbar anchorOrigin={{vertical: 'top', horizontal: 'center'}} autoHideDuration={4000} open={errorMessage.length > 0} onClose={()=> setErrorMessage('')}>
          <Alert variant='filled' severity="error" sx={{width: '100%'}}>{errorMessage}</Alert>
        </Snackbar>
      </Paper>
    </Container>
  );
}

export default Login;
