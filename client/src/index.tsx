import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';

import CssBaseline from '@mui/material/CssBaseline';
import {ThemeProvider} from '@mui/material/styles';
import theme from './ui/theme';

import './style/global.css';

import React from 'react';
import ReactDOM from 'react-dom';

import Config from './config';
import Application from './ui/appllication';


const config = new Config(process.env.REACT_APP_API_URL? process.env.REACT_APP_API_URL : '/api');

ReactDOM.render(
    <React.StrictMode>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Application config={config}/>
      </ThemeProvider>
    </React.StrictMode>,
    document.getElementById('root'),
);
