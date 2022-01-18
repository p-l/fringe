import axios from 'axios';
import {UserAuth} from '../../types/user-auth';

class AuthService {
  apiRootURL: string;

  public constructor() {
    this.apiRootURL = '/api/';
  }

  loginApiURL() :string {
    return this.apiRootURL + (this.apiRootURL.slice(-1) == '/' ? '' : '/') + 'auth/';
  }

  login(googleToken: string, googleTokenType: string, callback: (success: boolean, auth: UserAuth | null) => void) : void {
    axios.post(this.loginApiURL(), {access_token: googleToken, token_type: googleTokenType}).then( (response) => {
      console.debug(`auth service returned code:${response.status}`);
      if (!('token_type' in response.data && 'token' in response.data)) {
        console.warn(`invalid response from authentication API; missing token or token_type (code ${response.status})`);

        callback(false, null);
        return;
      }

      const auth = new UserAuth(response.data['token_type'], response.data['token']);
      console.debug(`authentication successful.`);
      callback(true, auth);
    }).catch((error) => {
      console.warn(`unable to authenticate to ${this.loginApiURL()}: ${error}`);
      callback(false, null);
    });
  }

  logout(callback: VoidFunction) : void {
    callback();
  }
}

const authService = new AuthService();
function useAuthService() : AuthService {
  return authService;
}

export default AuthService;
export {AuthService, useAuthService};

