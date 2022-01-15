export class UserAuth {
  token: string;
  tokenType: string;

  public constructor(token: string, tokenType: string) {
    this.token = token;
    this.tokenType = tokenType;
  }
}
