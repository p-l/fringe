import React from 'react';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import LightBulbIcon from '@mui/icons-material/Lightbulb';
import Config from '../../config';


interface AppProps {
    config: Config
}

function InfoLine(props: AppProps) {
  return (
    <Typography sx={{mt: 6, mb: 3}} color="text.secondary">
      <LightBulbIcon sx={{mr: 1, verticalAlign: 'middle'}} />
      API URL: {props.config.apiURL}
    </Typography>
  );
}

function Application(props: AppProps) {
  return (
    <Container maxWidth="sm">
      <Box sx={{my: 4}}>
        <Typography variant="h4" component="h1" gutterBottom>
          Client ready to implement
        </Typography>
        <InfoLine config={props.config}/>
      </Box>
    </Container>
  );
}

export default Application;
