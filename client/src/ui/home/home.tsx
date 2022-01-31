import React from 'react';
import Container from '@mui/material/Container';

import Config from '../../services/config';
import Me from '../me';


function Home({config}: {config: Config}) {
  return (
    <Container maxWidth="sm">
      <Me />
    </Container>
  );
}

export default Home;
