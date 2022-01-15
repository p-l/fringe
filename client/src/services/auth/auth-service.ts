import axios from 'axios';
import {UserAuth} from '../../types/user-auth';

class AuthService {
  authApiRootURL: string;

  public constructor() {
    this.authApiRootURL = '/api/auth';
  }

  login(googleToken: string, googleTokenType: string, callback: (success: boolean, auth: UserAuth | null) => void) : void {
    console.log('authService.login!');

    axios.post(this.authApiRootURL, {access_token: googleToken, token_type: googleTokenType}).then( (response) => {
      if (response.status != 200) {
        console.log(`Authentication failed (code ${response.status})`);

        callback(false, null);
        return;
      }

      if (!('token_type' in response.data && 'token' in response.data)) {
        console.log(`Invalid response from server. Missing token or token_type (code ${response.status})`);

        callback(false, null);
        return;
      }

      const auth = new UserAuth(response.data['token_type'], response.data['token']);
      callback(true, auth);
    }).catch((error) => {
      console.log(`unable to authenticate to fringe: ${error}`);
      callback(false, null);
    });
  }

  logout(callback: VoidFunction) : void {
    callback();
  }
}

const authService = new AuthService();
export function useAuthService() : AuthService {
  return authService;
}

export default AuthService;
export {AuthService};

