
// eslint-disable-next-line no-unused-vars
export enum UserAuthRole {unknown, user, admin}

export class UserAuth {
  token: string;
  tokenType: string;
  expires: number;
  private _role: UserAuthRole;
  private _roleString: string;

  public constructor(tokenType: string, token: string, expires: number, role: string) {
    this._roleString = 'unknown';
    this._role = UserAuthRole.unknown;
    this.token = token;
    this.tokenType = tokenType;
    this.expires = expires;

    this.roleString = role;
  }

  authorizationString(): string {
    return `${this.tokenType} ${this.token}`;
  }

  isExpired() : boolean {
    const now = Date.now();
    const expired = (this.expires - now <= 0);
    console.debug(`⏰ Auth token ${expired ? 'is expired' : 'is valid'} (expiry: ${new Date(this.expires).toDateString()} ${new Date(this.expires).toTimeString()})`);
    return expired;
  }

  public get roleString() : string {
    return this._roleString;
  }

  public get role() : UserAuthRole {
    return this._role;
  }

  public set roleString(roleString : string) {
    this._roleString = roleString;
    switch (roleString) {
      case 'admin':
        this.role = UserAuthRole.admin;
        break;
      case 'user':
        this.role = UserAuthRole.user;
        break;
      default:
        this.role = UserAuthRole.unknown;
        break;
    }
  }

  public set role(role: UserAuthRole) {
    this._role = role;
    switch (role) {
      case UserAuthRole.admin:
        this._roleString = 'admin';
        break;
      case UserAuthRole.user:
        this._roleString = 'user';
        break;
      case UserAuthRole.unknown:
        this._roleString = 'unknown';
        break;
    }
  }
}
