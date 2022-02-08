import axios from 'axios';

class Config {
  state: {
    loaded: boolean;
    error: Error | null;
  };

  apiRootURL: string;
  googleClientID: string;

  public constructor() {
    this.state = {
      loaded: false,
      error: null,
    };

    this.apiRootURL = 'https://'+window.location.host+'/api/';
    this.googleClientID = '';
  }

  configApiURL() : string {
    return this.apiRootURL + (this.apiRootURL.slice(-1) == '/' ? '' : '/') + 'config/';
  }

  waitForConfigFromAPI(loaded: (success: boolean, config: Config) => void) {
    console.debug('⚙️ Getting client configuration from: ' + this.configApiURL());
    axios.get(this.configApiURL()).then((r) => {
      // Minimal keys required for config to be deemed valid
      if (!r.data['google_client_id']) {
        this.state.loaded = false;
        this.state.error = new Error('Missing required configuration from API');
        loaded(false, this);

        return;
      }

      this.state.loaded = true;
      this.state.error = null;
      this.googleClientID = r.data['google_client_id'];

      console.debug('⚙️ Loaded config from:' + this.configApiURL());
      loaded(true, this);
    }).catch((e) => {
      this.state.loaded = false;
      this.state.error = e;

      console.warn(`⚙️ Config failed to load from ${this.configApiURL()}: ${e}`); // e can be anything, really
      loaded(false, this);
    });
  }
}

// tslint:disable-next-line no-var-requires prefer-template
export default Config;

