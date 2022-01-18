export class UserAuth {
  token: string;
  tokenType: string;

  public constructor(tokenType: string, token: string) {
    this.token = token;
    this.tokenType = tokenType;
  }

  authorizationString(): string {
    return `${this.tokenType} ${this.token}`;
  }
}
