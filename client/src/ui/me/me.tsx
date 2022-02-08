import React from 'react';
import {AccessTimeFilled, LockRounded, PasswordRounded} from '@mui/icons-material';
import {Alert, Avatar, Box, Button, CircularProgress, Container, Dialog, DialogActions, DialogContent, DialogTitle, List, ListItem, ListItemIcon, ListItemText, Paper, Snackbar, Typography} from '@mui/material';
import PasswordField from '../@components/password-field';

import useMountEffect from '../@hooks/use-mount';
import {User} from '../../models/user';
import {useUserService} from '../../services/user/user-service';
import theme from '../theme';

// eslint-disable-next-line no-unused-vars
enum PasswordState {NoPassword, FetchingPassword, HavePassword}

function Me() {
  const userService = useUserService();
  const [currentUser, setCurrentUser] = React.useState<User|null>(null);
  const [passwordState, setPasswordState] = React.useState<PasswordState>(PasswordState.NoPassword);
  const [errorMessage, setErrorMessage] = React.useState<string>('');
  const [warningModalOpen, setWarningModalOpen] = React.useState(false);

  const openWarningModal = () => setWarningModalOpen(true);
  const cancelWarningModal = () => setWarningModalOpen(false);
  const fetchUserPassword = () => {
    userService.renewMyPassword((user) => {
      if (user != null) {
        setCurrentUser(user);
        setPasswordState(PasswordState.HavePassword);
      } else {
        setPasswordState(PasswordState.NoPassword);
        setErrorMessage('Could not retrieve new password from server');
      }
    });
  };

  const confirmWarningModal = () => {
    setWarningModalOpen(false);
    setPasswordState(PasswordState.FetchingPassword);
    fetchUserPassword();
  };

  useMountEffect(() => {
    userService.me((user) => {
      setCurrentUser(user);
      if (user != null && user.password != null && user.password.length > 0) {
        setPasswordState(PasswordState.HavePassword);
      }
    });
  });

  return (
    <Container component="main" maxWidth="sm">
      {/* Error Message */}
      <Snackbar autoHideDuration={4000} open={errorMessage.length > 0} onClose={()=> setErrorMessage('')}>
        <Alert severity="error" sx={{width: '100%'}}>{errorMessage}</Alert>
      </Snackbar>
      {/* Confirmation Dialog */}
      <Dialog open={warningModalOpen}>
        <DialogTitle>Replacing Password</DialogTitle>
        <DialogContent>
          Creating a new password will remove the previous password. Only one password can exist at a time.
        </DialogContent>
        <DialogActions>
          <Button onClick={cancelWarningModal}>Cancel</Button>
          <Button onClick={confirmWarningModal} color="error">Understood</Button>
        </DialogActions>
      </Dialog>

      <Box sx={{marginTop: 5, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
        <Avatar
          sx={{bgcolor: theme.palette.primary.main, width: 96, height: 96}}
          alt={ currentUser?.email }
          src={ currentUser?.picture }
        />
        <Typography component="h1" variant="h5" sx={{p: 1}}>
          {currentUser?.name}
        </Typography>
      </Box>
      <Paper variant="outlined" sx={{marginTop: 4}}>
        <List>
          <ListItem>
            <ListItemIcon>
              <AccessTimeFilled />
            </ListItemIcon>
            <ListItemText
              primary="Last Seen"
              secondary={currentUser == null ? `...` : currentUser.lastSeen()}
            />
          </ListItem>
          <ListItem >
            <ListItemIcon>
              <LockRounded />
            </ListItemIcon>
            <ListItemText
              primary="Password Age"
              secondary={currentUser == null ? `...` : `${currentUser.passwordAgeInDays()} day old`}
            />
          </ListItem>
        </List>
        <Box sx={{p: 2, display: 'flex', flexDirection: 'column', alignItems: 'right'}}>
          { passwordState != PasswordState.HavePassword && (
            <Button
              onClick={openWarningModal}
              variant="contained"
              startIcon={ passwordState == PasswordState.FetchingPassword ? <CircularProgress size={16} />: <PasswordRounded />}
              disabled={passwordState != PasswordState.NoPassword}>
            New Password
            </Button>
          )}
          { passwordState == PasswordState.HavePassword && (
            <PasswordField loading={false} password={currentUser?.password} sx={{width: 1}}/>
          )}
        </Box>
      </Paper>
    </Container>
  );
}

export default Me;
