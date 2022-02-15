import axios from 'axios';
import {User} from '../../models/user';

class UserService {
  apiRootURL: string;

  public constructor() {
    this.apiRootURL = `https://${window.location.host}/api/`;
  }

  userApiURL() :string {
    return this.apiRootURL + (this.apiRootURL.slice(-1) == '/' ? '' : '/') + 'users/';
  }

  private static createUserFromResponseData(data : any) : User|null {
    const requiredFields = ['email', 'name', 'picture', 'last_seen_at', 'password_updated_at'];
    for (const field of requiredFields) {
      if (!(field in data)) {
        return null;
      }
    }

    return new User(data['email'], data['name'], data['picture'], data['last_seen_at'], data['password_updated_at'], data['password']);
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
    this.renewPassword('me', callback);
  }

  renewPassword(email:string, callback:(user: User|null) => void) {
    const renewURL = `${this.userApiURL()}${encodeURIComponent(email)}/renew/`;
    axios.get(renewURL).then((response) => {
      const user = UserService.createUserFromResponseData(response.data);
      if (user == null) {
        console.warn(`Invalid user response from user API at ${renewURL}`);
        callback(null);
      } else if (user.password == null || user.password.length < 1) {
        console.warn(`No password provided in password renew API at ${renewURL}`);
        callback(null);
      } else {
        callback(user);
      }
    }).catch((error) => {
      console.warn(`Unable to renew password from ${renewURL}: ${error}`);
      callback(null);
    });
  }

  findAllUsers(searchQuery : string, page : number, perPage : number, callback : (users: User[], success: boolean) => void) : void {
    const params = {
      'per_page': perPage,
      'page': page,
      'search': searchQuery,
    };
    axios.get(this.userApiURL(), {params: params} ).then((response) => {
      console.debug(response);

      const users : User[] = [];
      if (response.data instanceof Array) {
        for (let i = 0; i < response.data.length; i++) {
          const userData = response.data[i];
          const user = UserService.createUserFromResponseData(userData);
          if (user != null) {
            users.push(user);
          }
        }
      }
      callback(users, true);
    }).catch((error) => {
      console.warn(`Unable to retrieve user list from ${this.userApiURL()}: ${error}`);
      callback([], false);
    });
  }

  create(email: string, name:string|null, callback: (resultText: string, user: User|null, )=>void) {
    axios.post(this.userApiURL(), {'email': email, 'name': name}).then((response) => {
      let user : User|null = null;
      let result : string = 'failed';

      if (response.data && 'result' in response.data) {
        result = response.data['result'];
        if (result == 'success' && 'user' in response.data) {
          const userData = response.data['user'];
          user = UserService.createUserFromResponseData(userData);
        }
      }

      callback(result, user);
    }).catch((error) => {
      console.warn(`Failed to created user at ${this.userApiURL()}: ${error}`);
      callback('failed', null);
    });
  }

  delete(email: string, callback: (resultText: string)=>void) {
    const deleteURL = `${this.userApiURL()}${encodeURIComponent(email)}/`;
    axios.delete(deleteURL).then((response) => {
      let result : string = 'failed';
      if (response.data && 'result' in response.data) {
        result = response.data['result'];
      }
      callback(result);
    }).catch((error) => {
      console.warn(`Failed to delete user at ${deleteURL}: ${error}`);
      callback('failed');
    });
  }
}

const userService = new UserService();
function useUserService() : UserService {
  return userService;
}

export default UserService;
export {UserService, useUserService};

