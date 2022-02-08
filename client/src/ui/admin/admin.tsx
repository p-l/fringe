import React from 'react';
import {ClearRounded, DeleteRounded, MoreRounded, PasswordRounded, PersonAddRounded, PersonSearchRounded} from '@mui/icons-material';
import {Avatar, Box, Button, Container, Dialog, IconButton, Input, InputAdornment, Table, TableContainer, TableRow, TableCell, TableBody, CircularProgress, DialogTitle, DialogContent, DialogActions, TableHead} from '@mui/material';
import {User} from '../../models/user';
import PasswordField from '../@components/password-field';
import theme from '../theme';

const appendMoreUsers = (currentList: User[], amount: number) : User[] => {
  const newUsers:User[] = [];

  for (let i = currentList.length; i < currentList.length+amount; i++) {
    const user = new User(
        'user-'+(i+1)+'@domain.com',
        'User'+(i+1)+' WithALastName',
        'https://picsum.photos/96/96',
        (Date.now()/1000)-(i*60*60),
        (Date.now()/1000)-(i*24*60*60),
        null);
    newUsers.push(user);
  }
  return currentList.concat(newUsers);
};

let searchTimer = setTimeout(()=>{}, 1000);

function Admin() {
  const [users, setUsers] = React.useState<User[]>(appendMoreUsers([], 20));
  const [loading, setLoading] = React.useState<boolean>(false);
  const [userDelete, setUserDelete] = React.useState<User|null>(null);
  const [userReset, setUserReset] = React.useState<User|null>(null);
  const [doUserReset, setDoUserReset] = React.useState<boolean>(false);
  const [userResetPassword, setUserResetPassword] = React.useState<string|null>( null);
  const [search, setSearch] = React.useState<string>('');

  const loadMoreUsers = () => {
    setLoading(true);
    setTimeout(()=>{
      const newUserList = appendMoreUsers(users, 20);
      setUsers(newUserList);
      setLoading(false);
    }, 500);
  };

  const deleteButtonPressed = (user: User) => {
    setUserDelete(user);
    setUserReset(null);
    setDoUserReset(false);
    setUserResetPassword(null);
  };

  const cancelUserDelete = () => {
    setUserDelete(null);
    setUserReset(null);
    setDoUserReset(false);
    setUserResetPassword(null);
  };

  const confirmUserDelete = () => {
    setUsers(users.filter((user) => user.email != userDelete?.email));
    setUserDelete(null);
    setUserReset(null);
    setDoUserReset(false);
    setUserResetPassword(null);
  };

  const resetUserPassword = (user: User) => {
    setUserDelete(null);
    setUserReset(user);
    setDoUserReset(false);
    setUserResetPassword(null);
  };

  const confirmUserReset = () => {
    setUserDelete(null);
    setDoUserReset(true);
    setTimeout(() => {
      setUserResetPassword('SOME_RANDOM_PASSWORD_HERE');
    }, 2000);
  };

  const cancelUserReset = () => {
    setUserDelete(null);
    setUserReset(null);
    setDoUserReset(false);
    setUserResetPassword(null);
  };

  const closePasswordDialog = () => {
    setUserDelete(null);
    setUserReset(null);
    setDoUserReset(false);
    setUserResetPassword(null);
  };

  const doSearch = (searchString: string, delayed: boolean = true) => {
    setSearch(searchString);
    clearTimeout(searchTimer);
    if (searchString != '') {
      searchTimer = setTimeout(() => {
        alert('🔍 ' + searchString);
      }, delayed ? 800 : 0);
    }
  };


  return (
    <Container component="main" >
      {/* Confirm Delete */}
      <Dialog open={userDelete != null}>
        <DialogTitle>Delete {userDelete?.email}?</DialogTitle>
        <DialogContent>Remove user from the database. This does not prevent the user from enrolling again later.</DialogContent>
        <DialogActions>
          <Button onClick={cancelUserDelete}>Cancel</Button>
          <Button onClick={confirmUserDelete} color="error">Delete</Button>
        </DialogActions>
      </Dialog>
      {/* Confirm Password Reset */}
      <Dialog open={userReset != null && !doUserReset}>
        <DialogTitle>Reset {userReset?.email} password?</DialogTitle>
        <DialogContent>This will change the current user password for a new password. It will invalid current user password.</DialogContent>
        <DialogActions>
          <Button onClick={cancelUserReset}>Cancel</Button>
          <Button onClick={confirmUserReset} color="error">Reset</Button>
        </DialogActions>
      </Dialog>
      {/* Show new Password */}
      <Dialog open={userReset != null && doUserReset}>
        <DialogTitle>New password for {userReset?.email}</DialogTitle>
        <DialogContent>
          Make sure to copy the password below. Once the dialog closed it cannot be retrieved again.
        </DialogContent>
        <DialogContent sx={{topMargin: 0}}>
          <PasswordField loading={userResetPassword == null} password={userResetPassword} sx={{width: 1}} />
        </DialogContent>
        <DialogActions>
          <Button onClick={closePasswordDialog}>Close</Button>
        </DialogActions>
      </Dialog>
      {/* Search and Add */}
      <Box sx={{marginTop: 2, display: 'flex', alignItems: 'center'}}>
        <Input
          id='user-search'
          sx={{width: '100%'}}
          value={search}
          onChange={(event) => doSearch(event.target.value)}
          onKeyPress={(event) => {
            if (event.key === 'Enter') {
              const thisInput = document.getElementById('user-search') as HTMLInputElement;
              doSearch(thisInput.value, false);
            }
          }}
          startAdornment={<InputAdornment position='start'><IconButton onClick={() => doSearch(search, false)}><PersonSearchRounded /></IconButton></InputAdornment> }
          endAdornment={(search.length > 0) && <InputAdornment position='end'><IconButton onClick={ () => doSearch('', false)}><ClearRounded /></IconButton></InputAdornment>}
          placeholder='Find by name or email'
        />
        <IconButton><PersonAddRounded /></IconButton>
      </Box>
      {/* User table */}
      <TableContainer>
        <Table aria-label="user table">
          <TableHead sx={{display: {xs: 'none', sm: 'none', md: 'table-header-group', lg: 'table-header-group', xl: 'table-header-group'}}}>
            <TableRow>
              <TableCell sx={{paddingLeft: 0}}>&nbsp;</TableCell>
              <TableCell sx={{paddingLeft: 0}} align="left">Email</TableCell>
              <TableCell align="left" sx={{display: {xs: 'none', sm: 'table-cell', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>Name</TableCell>
              <TableCell align="left" sx={{display: {xs: 'none', sm: 'none', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>Last Seen</TableCell>
              <TableCell align="left" sx={{display: {xs: 'none', sm: 'none', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>Last Password Change</TableCell>
              <TableCell sx={{paddingRight: 0}} align="right">&nbsp;</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {users.map((user) => (
              <TableRow
                key={user.email}
                sx={{'&:last-child td, &:last-child th': {border: 0}}}
              >
                <TableCell sx={{paddingLeft: 0}} >
                  <Avatar
                    sx={{bgcolor: theme.palette.primary.main}}
                    alt={ user.email }
                    src={ user.picture } />
                </TableCell>
                <TableCell sx={{paddingLeft: 0}} align="left">{user.email}</TableCell>
                <TableCell align="left" sx={{display: {xs: 'none', sm: 'table-cell', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>{user.name}</TableCell>
                <TableCell align="left" sx={{display: {xs: 'none', sm: 'none', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>{user.lastSeen()}</TableCell>
                <TableCell align="left" sx={{display: {xs: 'none', sm: 'none', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>{user.passwordAgeInDays()} days ago</TableCell>
                <TableCell sx={{paddingRight: 0}} align="right">
                  <IconButton onClick={() => resetUserPassword(user)}>
                    <PasswordRounded />
                  </IconButton>
                  <IconButton onClick={() => deleteButtonPressed(user)}>
                    <DeleteRounded />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      {/* Load More... */}
      <Box sx={{marginTop: 2, marginBottom: 2}}>
        <Button
          variant="outlined"
          sx={{width: 1}}
          onClick={loadMoreUsers}
          disabled={loading}
          startIcon={loading ? <CircularProgress sx={{height: 18, width: 18}} /> : <MoreRounded />}
        >Load More...</Button>
      </Box>
    </Container>
  );
}

export default Admin;
