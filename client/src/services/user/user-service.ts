import axios from 'axios';
import {User} from '../../models/user';

class UserService {
  apiRootURL: string;

  public constructor() {
    this.apiRootURL = 'https://'+window.location.host+'/api/';
  }

  userApiURL() :string {
    return this.apiRootURL + (this.apiRootURL.slice(-1) == '/' ? '' : '/') + 'users/';
  }

  private static createUserFromResponseData(data : any) : User|null {
    if ('email' in data && 'last_seen_at' in data) {
      return new User(data['email'], data['name'], data['picture'], data['last_seen_at'], data['password_updated_at'], data['password']);
    }

    return null;
  }

  me(callback: (user: User|null) => void) : void {
    axios.get(this.userApiURL()+'me/').then((response) => {
      const user = UserService.createUserFromResponseData(response.data);
      if (user == null) {
        console.warn(`Invalid user response from user API at ${this.userApiURL()+'me/'}`);
      }

      callback(user);
    }).catch((error) => {
      console.warn(`Unable to retrieve user from ${this.userApiURL()+'me/'}: ${error}`);
      callback(null);
    });
  }

  renewMyPassword(callback: (user: User|null) => void) : void {
    axios.get(this.userApiURL()+'me/renew/').then((response) => {
      const user = UserService.createUserFromResponseData(response.data);
      if (user == null) {
        console.warn(`Invalid user response from user API at ${this.userApiURL()+'me/renew/'}`);
        callback(null);
      } else if (user.password == null || user.password.length < 1) {
        console.warn(`No password provided in password renew API at ${this.userApiURL()+'me/renew/'}`);
        callback(null);
      } else {
        callback(user);
      }
    }).catch((error) => {
      console.warn(`Unable to renew password from ${this.userApiURL()+'me/renew/'}: ${error}`);
      callback(null);
    });
  }
}

const userService = new UserService();
function useUserService() : UserService {
  return userService;
}

export default UserService;
export {UserService, useUserService};

