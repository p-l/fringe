import React from 'react';
import {render, screen} from '@testing-library/react';
import Application from './application';
import Config from '../../config';

test('renders learn react link', () => {
  const config = new Config();
  config.state.loaded = true;
  config.state.error = null;
  config.apiURL = 'file:///';
  config.googleClientID = 'client_id';

  render(<Application config={config} />);
  const linkElement = screen.getByText(/forget react/i);
  expect(linkElement).toBeInTheDocument();
});
