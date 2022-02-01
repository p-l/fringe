import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';
import Config from './config';

describe('Config', () => {
  let mock : MockAdapter;

  beforeAll(() => {
    mock = new MockAdapter(axios);
  });

  afterEach(() => {
    mock.reset();
  });

  it('calls back upon successful retrieval of config', async () => {
    const config = new Config();
    mock.onGet(config.configApiURL()).reply(200, {
      'google_client_id': '123',
    });
    config.waitForConfigFromAPI( (success, config) => {
      expect(success).toBe(true);
      expect(config.state.loaded).toBe(true);
      expect(config.state.error).toBeNull();
      expect(config.googleClientID).toBe('123');
    });
  });

  it('calls back upon incomplete config', async () => {
    const config = new Config();
    mock.onGet(config.configApiURL()).reply(200, {});

    config.waitForConfigFromAPI( (success, config) => {
      expect(success).toBe(false);
      expect(config.state.loaded).toBe(false);
      expect(config.state.error).not.toBeNull();
      expect(config.googleClientID).toBe('');
    });
  });

  it('returns invalid config on invalid api config', async () => {
    const config = new Config();
    mock.onGet(config.configApiURL()).reply(500);

    config.waitForConfigFromAPI((success, config) => {
      expect(success).toBe(false);
      expect(config.state.loaded).toBe(false);
      expect(config.state.error).not.toBeNull();
      expect(config.googleClientID).toBe('');
    });
  });

  it('does not double slash config URL', async () => {
    const config = new Config();
    expect(config.apiRootURL).toContain('/api/');
    expect(config.configApiURL()).toContain('/api/config/');

    config.apiRootURL = '/test/api';
    expect(config.apiRootURL).toContain('/test/api');
    expect(config.configApiURL()).toContain('/test/api/config/');
  });
});
