import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';

import CssBaseline from '@mui/material/CssBaseline';
import {ThemeProvider} from '@mui/material/styles';
import theme from './ui/theme';

import React from 'react';
import ReactDOM from 'react-dom';
import {BrowserRouter as Router} from 'react-router-dom';

import Application from './ui/appllication';
import Config from './config';

const apiRootURL = (process.env.REACT_APP_API_URL? process.env.REACT_APP_API_URL : '/api');

new Config().waitForConfigFromAPI(apiRootURL, (config : Config) => {
  ReactDOM.render(
      <React.StrictMode>
        <ThemeProvider theme={theme}>
          <CssBaseline/>
          <Router>
            <Application config={config}/>
          </Router>
        </ThemeProvider>
      </React.StrictMode>,
      document.getElementById('root'),
  );
});
