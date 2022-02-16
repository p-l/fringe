import {differenceInDays} from 'date-fns';
import React from 'react';
import {AccessTimeFilled, LockRounded, PasswordRounded} from '@mui/icons-material';
import {Alert, Avatar, Box, Button, CircularProgress, Container, Dialog, DialogActions, DialogContent, DialogTitle, List, ListItem, ListItemIcon, ListItemText, Paper, Snackbar, Typography} from '@mui/material';
import {Trans, useTranslation} from 'react-i18next';

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
  const {t} = useTranslation();

  const openWarningModal = () => setWarningModalOpen(true);
  const cancelWarningModal = () => setWarningModalOpen(false);
  const fetchUserPassword = () => {
    userService.renewMyPassword((user) => {
      if (user != null) {
        setCurrentUser(user);
        setPasswordState(PasswordState.HavePassword);
      } else {
        setPasswordState(PasswordState.NoPassword);
        setErrorMessage(t('me.errorFailedToGetPassword'));
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
      <Snackbar anchorOrigin={{vertical: 'bottom', horizontal: 'center'}} autoHideDuration={4000} open={errorMessage.length > 0} onClose={()=> setErrorMessage('')}>
        <Alert variant='filled' severity="error" sx={{width: '100%'}}>{errorMessage}</Alert>
      </Snackbar>
      {/* Confirmation Dialog */}
      <Dialog open={warningModalOpen}>
        <DialogTitle><Trans i18nKey='me.renewDialogTitle' /></DialogTitle>
        <DialogContent><Trans i18nKey='me.renewDialogInstruction' /></DialogContent>
        <DialogActions>
          <Button onClick={cancelWarningModal}><Trans i18nKey='actions.cancel' /></Button>
          <Button onClick={confirmWarningModal} color="error"><Trans i18nKey='actions.understood' /></Button>
        </DialogActions>
      </Dialog>
      {/* User profile */}
      <Box sx={{marginTop: 5, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
        <Avatar
          sx={{bgcolor: theme.palette.primary.main, width: 96, height: 96}}
          alt={ currentUser?.email.toLocaleUpperCase() }
          src={ currentUser?.picture }
        />
        <Typography component="h1" variant="h5" sx={{p: 1}}>
          {currentUser?.name}
        </Typography>
      </Box>
      <Paper variant="outlined" sx={{marginTop: 4}}>
        <List>
          {/* Last Seen */}
          <ListItem>
            <ListItemIcon>
              <AccessTimeFilled />
            </ListItemIcon>
            <ListItemText
              primary={t('me.lastSeen')}
              secondary={currentUser == null ? `...` : t('me.lastSeenDate', {lastSeenDate: currentUser.lastSeenAt, formatParams: {lastSeenDate: {weekday: 'long', year: 'numeric', month: 'long', day: 'numeric', hour: 'numeric', minute: 'numeric', second: 'numeric', hour12: false, timeZoneName: 'short'}}})}
            />
          </ListItem>
          {/* Password Age */}
          <ListItem >
            <ListItemIcon>
              <LockRounded />
            </ListItemIcon>
            <ListItemText
              primary={t('me.passwordAge')}
              secondary={currentUser == null ? `...` : t('me.passwordAgeRelative', {count: differenceInDays(currentUser.passwordUpdatedAt, Date.now()), passwordAge: differenceInDays(currentUser.passwordUpdatedAt, Date.now())})}
            />
          </ListItem>
        </List>
        {/* Password */}
        <Box sx={{p: 2, display: 'flex', flexDirection: 'column', alignItems: 'right'}}>
          { passwordState != PasswordState.HavePassword && (
            <Button
              onClick={openWarningModal}
              variant="contained"
              startIcon={ passwordState == PasswordState.FetchingPassword ? <CircularProgress size={16} />: <PasswordRounded />}
              disabled={passwordState != PasswordState.NoPassword}>
              <Trans i18nKey='me.newPassword' />
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
