import axios from 'axios';

class Config {
  apiURL: string;
  loaded: boolean;
  google_client_id: string;

  constructor(apiURL: string) {
    this.apiURL = apiURL || '';
    this.google_client_id = '';
    this.loaded = false;

    if (apiURL != '') {
      const configURL = this.apiURL + '/config/';
      console.debug('Getting client configuration from:' + configURL);
      axios.get(configURL).then((r) => {
        if (!r.data['google_client_id']) return;

        console.debug('Valid config from:' + configURL);
        this.google_client_id = r.data['google_client_id'];
        this.loaded = true;
      }).catch((e) => {
        console.log(e); // e can be anything, really.
      });
    }
  }
}


// tslint:disable-next-line no-var-requires prefer-template
export default Config;

