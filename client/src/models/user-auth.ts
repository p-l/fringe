export class UserAuth {
  token: string;
  tokenType: string;
  expires: number;

  public constructor(tokenType: string, token: string, expires: number) {
    this.token = token;
    this.tokenType = tokenType;
    this.expires = expires;
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
}
