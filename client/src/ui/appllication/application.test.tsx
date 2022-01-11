import React from 'react';
import {render, screen} from '@testing-library/react';
import Application from './application';
import Config from '../../config/';

test('renders learn react link', () => {
  const config = new Config('');
  render(<Application config={config}/>);
  const linkElement = screen.getByText(/forget react/i);
  expect(linkElement).toBeInTheDocument();
});
