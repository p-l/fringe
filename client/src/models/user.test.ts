import {User} from './user';

describe('User', () => {
  it('keeps password null', async () => {
    const user = new User('email@email.com', 'name', 'https://picture.url/somewhere', 0, 0, null);
    expect(user.password).toBeNull();
  });

  it('set password to null when string is empty', async () => {
    const user = new User('email@email.com', 'name', 'https://picture.url/somewhere', 0, 0, '');
    expect(user.password).toBeNull();
  });
});
