import axios from 'axios';
import {UserAuth} from '../../models/user-auth';

class AuthService {
  apiRootURL: string;
  private _userAuth: UserAuth|null;

  public constructor() {
    this.apiRootURL = `https://${window.location.host}/api/`;
    this._userAuth = null;

    const localToken = localStorage.getItem('token');
    const localTokenType = localStorage.getItem('token_type');
    const localTokenExpiresString = localStorage.getItem('token_expires_at');
    let localTokenRole = localStorage.getItem('token_role');
    if (localTokenRole == null) {
      localTokenRole = 'unknown';
    }
    if ( localToken && localTokenType && localTokenExpiresString) {
      const localTokenExpires = Number(localTokenExpiresString);
      const auth = new UserAuth(localTokenType, localToken, localTokenExpires, localTokenRole);
      if (!auth.isExpired()) {
        this._userAuth = auth;
      }
    }
  }

  public get currentUserAuth() : UserAuth|null {
    return this._userAuth;
  }

  public set currentUserAuth(auth: UserAuth|null) {
    if (auth == null || auth.isExpired()) {
      console.debug(`🛂 Removing auth token from storage`);
      localStorage.removeItem('token');
      localStorage.removeItem('token_type');
      localStorage.removeItem('token_expires_at');
      localStorage.removeItem('token_role');
      this._userAuth = null;

      return;
    }

    localStorage.setItem('token_type', auth.tokenType);
    localStorage.setItem('token', auth.token);
    localStorage.setItem('token_expires_at', auth.expires.toString());
    localStorage.setItem('token_role', auth.roleString);
    this._userAuth = auth;
  }

  public loginApiURL() :string {
    return this.apiRootURL + (this.apiRootURL.slice(-1) == '/' ? '' : '/') + 'auth/';
  }

  public login(googleToken: string, googleTokenType: string, callback: (success: boolean, auth: UserAuth | null) => void) : void {
    axios.post(this.loginApiURL(), {access_token: googleToken, token_type: googleTokenType}).then( (response) => {
      console.debug(`🛂 Auth service returned code:${response.status}`);
      if (!('token_type' in response.data && 'token' in response.data)) {
        console.warn(`Invalid response from authentication API; missing token or token_type (code ${response.status})`);

        callback(false, null);
        return;
      }

      const durationInMilliseconds = Number(response.data['duration'])*1000;
      const expiry = Date.now()+durationInMilliseconds;

      const auth = new UserAuth(response.data['token_type'], response.data['token'], expiry, response.data['role']);
      this.currentUserAuth = auth;

      callback(true, auth);
    }).catch((error) => {
      console.warn(`Unable to authenticate to ${this.loginApiURL()}: ${error}`);
      callback(false, null);
    });
  }

  public logout(callback: VoidFunction) : void {
    this.currentUserAuth = null;
    callback();
  }
}

const authService = new AuthService();
function useAuthService() : AuthService {
  return authService;
}

axios.interceptors.request.use((config) => {
  if (config.url != null && config.headers != null) {
    if (config.url.startsWith(authService.apiRootURL)) {
      if (authService.currentUserAuth != null) {
        console.debug(`🛂 Adding Authorization headers to request: ${config.url}`);
        config.headers['Authorization'] = authService.currentUserAuth.authorizationString();
      }
    }
  }
  return config;
});

export default AuthService;
export {AuthService, useAuthService};

