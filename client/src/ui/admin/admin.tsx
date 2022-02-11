import React from 'react';
import {MoreRounded} from '@mui/icons-material';
import {Avatar, Box, Button, CircularProgress, Container, Stack, Table, TableContainer, TableRow, TableCell, TableBody, TableHead} from '@mui/material';

import {User} from '../../models/user';
import {useUserService} from '../../services/user/user-service';
import useMountEffect from '../@hooks/use-mount';
import UserAdd from './@components/user-add/user-add';
import UserDelete from './@components/user-delete';
import UserRenew from './@components/user-renew';
import UserSearch from './@components/user-search';

function Admin() {
  const [users, setUsers] = React.useState<User[]>([]);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [query, setQuery] = React.useState<string>('');
  const [page, setPage] = React.useState<number>(1);
  const [hasMore, setHasMore] = React.useState<boolean>(true);
  const userService = useUserService();

  useMountEffect(()=> {
    getUsers('', 1);
  });

  const getUsers = (query:string, page:number) => {
    setQuery(query);
    setPage(page);
    setLoading(true);
    const perPage = 20;
    userService.findAllUsers(query, page, perPage, (foundUsers, success) => {
      setLoading(false);
      if (success) {
        if (page <= 1) {
          setUsers(foundUsers);
        } else {
          setUsers(users.concat(foundUsers));
        }
        setHasMore(foundUsers.length == perPage);
      } else {
        console.warn('Failed to retrieve user list');
      }
    });
  };

  const loadMoreUsers = () => {
    getUsers(query, page+1);
  };

  const newSearch = (newQuery:string = '') => {
    setUsers([]);
    getUsers(newQuery, 1);
  };

  return (
    <Container component="main" >
      {/* Search and Add */}
      <Box sx={{marginTop: 2, display: 'flex', alignItems: 'center'}}>
        <UserSearch onSearch={newSearch} onClear={newSearch} delay={750} sx={{width: '100%'}} />
        <UserAdd onCreation={(newUser) => {
          const updatedUsers = [newUser].concat(users);
          setUsers(updatedUsers);
        }}/>
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
              <TableCell sx={{paddingRight: 0, maxWidth: 20}} align="right">&nbsp;</TableCell>
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
                    alt={ user.email.toLocaleUpperCase() }
                    src={ user.picture } />
                </TableCell>
                <TableCell sx={{paddingLeft: 0}} align="left">{user.email}</TableCell>
                <TableCell align="left" sx={{display: {xs: 'none', sm: 'table-cell', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>{user.name}</TableCell>
                <TableCell align="left" sx={{display: {xs: 'none', sm: 'none', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>{user.lastSeen()}</TableCell>
                <TableCell align="left" sx={{display: {xs: 'none', sm: 'none', md: 'table-cell', lg: 'table-cell', xl: 'table-cell'}, paddingLeft: 0}}>{user.passwordAgeInDays()} days ago</TableCell>
                <TableCell sx={{paddingRight: 0}} align="right">
                  <Stack direction="row" spacing={0} justifyContent="flex-end">
                    <UserDelete user={user} onDelete={(deletedUser) => {
                      const updatedUsers = users.filter((u) => u.email != deletedUser.email);
                      setUsers(updatedUsers);
                    }} />
                    <UserRenew user={user} />
                  </Stack>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      {/*  More... */}

      <Box sx={{marginTop: 2, marginBottom: 2}}>
        <Button
          variant="outlined"
          sx={{width: 1}}
          onClick={loadMoreUsers}
          disabled={loading || !hasMore}
          startIcon={loading ? <CircularProgress size={12} /> : <MoreRounded />}
        >{hasMore ? (`More Results`) : (`No More Users`)}</Button>
      </Box>
    </Container>
  );
}

export default Admin;
