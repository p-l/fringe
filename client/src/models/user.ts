export class User {
  email: string;
  name: string;
  picture: string;
  lastSeenAt: Date;
  passwordUpdatedAt: Date;
  password: string|null;


  public constructor(email: string, name: string, picture: string, unixLastSeenAt: number, unixPasswordUpdatedAt: number, password: string|null = null) {
    this.email = email;
    this.name = name;
    this.picture = picture;
    this.lastSeenAt = new Date(unixLastSeenAt * 1000);
    this.passwordUpdatedAt = new Date(unixPasswordUpdatedAt * 1000);
    if (password != null && password.length > 0) {
      this.password = password;
    } else {
      this.password = null;
    }
  }
}
