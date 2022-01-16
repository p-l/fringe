import axios from 'axios';
import Config from './config';

jest.mock('axios');

describe('waitForConfigFromAPI', () => {
  it('fetches successfully data from an API', async () => {
    const configData = {'google_client_id': '123'};
    const mockedGet = jest
        .spyOn(axios, 'get')
        .mockImplementation(() => Promise.resolve({data: configData}));


    const config = new Config();
    config.waitForConfigFromAPI('/api/', (config) => {
      expect(config.apiURL).toBe('/api/');
      expect(config.googleClientID).toBe('123');
    });

    expect(mockedGet).toHaveBeenCalledWith(
        `/api/config/`,
    );
  });

  it('fetches erroneously data from an API', async () => {

  });
});
