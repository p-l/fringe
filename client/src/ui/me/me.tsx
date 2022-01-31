import React from 'react';
import {AccessTimeFilled, CheckCircleOutlineRounded, CopyAllRounded, LockRounded, PasswordRounded, WarningOutlined} from '@mui/icons-material';
import {Alert, Avatar, Backdrop, Box, Button, CircularProgress, Container, Fade, FormControl, IconButton, Input, InputAdornment, InputLabel, List, ListItem, ListItemIcon, ListItemText, Modal, Paper, Snackbar, Stack, Typography} from '@mui/material';

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
  const [passwordCopied, setPasswordCopied] = React.useState<boolean>(false);
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
  const copyUserPassword = () => {
    if (currentUser != null && currentUser.password != null) {
      navigator.clipboard.writeText(currentUser.password).then(() => {
        setPasswordCopied(true);
      });
    }
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
      { /* Password Copy Confirmation */ }
      <Snackbar
        autoHideDuration={3000}
        open={passwordCopied}
        onClose={() => setPasswordCopied(false)}
        message="Password was copied to clipboard"
      />
      <Box sx={{marginTop: 5, display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
        <Avatar
          sx={{bgcolor: theme.palette.primary.main, width: 96, height: 96}}
          alt={ currentUser?.name }
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
              secondary={currentUser == null ? `...` : `${currentUser?.lastSeenAt.toDateString()} at ${currentUser?.lastSeenAt.toTimeString().split(' ')[0]}`}
            />
          </ListItem>
          <ListItem >
            <ListItemIcon>
              <LockRounded />
            </ListItemIcon>
            <ListItemText
              primary="Password Age"
              secondary={currentUser == null ? `...` : `${Math.ceil(Math.abs(Date.now() - currentUser?.passwordUpdatedAt.getTime())/(24*60*60*1000))} day old`}
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
            <FormControl variant="standard">
              <InputLabel htmlFor="fringe-user-password">Password</InputLabel>
              <Input
                id="fringe-user-password"
                readOnly={true}
                disabled={true}
                type="text"
                value={currentUser?.password}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="copy password to clipboard"
                      onClick={copyUserPassword}
                      edge="end"
                    >
                      { passwordCopied ? (<CheckCircleOutlineRounded />) : (<CopyAllRounded />)}
                    </IconButton>
                  </InputAdornment>
                }
              />
            </FormControl>
          )}

        </Box>
      </Paper>
      <Modal
        aria-labelledby="transition-modal-title"
        aria-describedby="transition-modal-description"
        open={warningModalOpen}
        onClose={cancelWarningModal}
        closeAfterTransition
        BackdropComponent={Backdrop}
        BackdropProps={{timeout: 500}}>
        <Fade in={warningModalOpen}>
          <Box sx={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            width: 400,
            bgcolor: 'background.paper',
            border: '2px solid #000',
            boxShadow: 24,
            p: 4}}>
            <Stack direction="row" alignItems="center" gap={1}>
              <WarningOutlined />
              <Typography id="transition-modal-title" variant="h6" component="h2">
                Replacing Password
              </Typography>
            </Stack>
            <Typography id="transition-modal-description" sx={{mt: 2}}>
              Creating a new password will remove the previous password. Only one password can exist at a time.
            </Typography>
            <Box sx={{marginTop: 2}}/>
            <Stack direction="row" justifyContent="flex-end" alignItems="center" spacing={2}>
              <Button onClick={cancelWarningModal} variant="outlined" color="error">Take me back</Button>
              <Button onClick={confirmWarningModal} variant="contained">Understood</Button>
            </Stack>
          </Box>
        </Fade>
      </Modal>


    </Container>


  );
}

export default Me;
