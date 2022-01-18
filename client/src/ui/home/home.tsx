import React from 'react';

import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import LightBulbIcon from '@mui/icons-material/Lightbulb';

import Config from '../../services/config';
import {GoogleLogout} from 'react-google-login';
import {useNavigate} from 'react-router-dom';
import {useAuth} from '../@contexts/auth';


function Home({config}: {config: Config}) {
  const auth = useAuth();
  const navigate = useNavigate();

  const googleLogoutSuccessHandler = () => {
    auth.logout(()=>{
      console.log('Navigate to /');
      navigate('/login', {replace: true});
    });
  };
  return (
    <Container maxWidth="sm">
      <Box sx={{my: 4}}>
        <Typography variant="h4" component="h1" gutterBottom>
          Client ready to implement
        </Typography>
        <Typography sx={{mt: 6, mb: 3}} color="text.secondary">
          <LightBulbIcon sx={{mr: 1, verticalAlign: 'middle'}} />
            Google Client ID: {config.googleClientID}
        </Typography>
        <GoogleLogout
          clientId={config.googleClientID}
          onLogoutSuccess={googleLogoutSuccessHandler}/>
      </Box>
    </Container>
  );
}

export default Home;
