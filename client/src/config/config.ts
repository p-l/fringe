import axios from 'axios';


class Config {
  state: {
    loaded: boolean;
    error: Error | null;
  };

  apiURL: string;
  googleClientID: string;

  public constructor() {
    this.state = {
      loaded: false,
      error: null,
    };

    this.apiURL = '';
    this.googleClientID = '';
  }

  waitForConfigFromAPI(apiRootURL: string, loaded: ConfigLoaded) {
    console.log('Loading configuration from: '+apiRootURL);
    const configURL = apiRootURL + '/config/';
    console.debug('Getting client configuration from: ' + apiRootURL);
    axios.get(configURL).then((r) => {
      // Minimal keys required for config to be deemed valid
      if (!r.data['google_client_id']) throw Error('Invalid configuration file. (Missing `google_client_id`)');

      console.debug('Valid config from:' + configURL);

      this.apiURL = apiRootURL;
      this.googleClientID = r.data['google_client_id'];

      loaded(this);
    }).catch((e) => {
      this.state.loaded = false;
      this.state.error = e;

      console.log('Config loading failed: '+e); // e can be anything, really.
      loaded(this);
    });
  }
}

type ConfigLoaded = (config: Config) => any;


// tslint:disable-next-line no-var-requires prefer-template
export default Config;

