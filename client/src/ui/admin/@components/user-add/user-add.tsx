import {PersonAddRounded} from '@mui/icons-material';
import {Alert, Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, SxProps, TextField, Theme, Typography} from '@mui/material';
import React from 'react';
import validator from 'validator';
import {User} from '../../../../models/user';
import {useUserService} from '../../../../services/user/user-service';
import PasswordDialog from '../password-dialog';

function UserAdd(props: {onCreation:(user: User) => void, sx?: SxProps<Theme>|undefined}) {
  const [creationDialogOpened, setCreationDialogOpened] = React.useState<boolean>(false);
  const [email, setEmail] = React.useState<string>('');
  const [name, setName] = React.useState<string>('');
  const [error, setError] = React.useState<string|null>(null);
  const [isEmailValid, setIsEmailValid] = React.useState<boolean>(false);
  const [creating, setCreating] = React.useState<boolean>(false);

  const [passwordDialogOpened, setPasswordDialogOpened] = React.useState<boolean>(false);
  const [passwordDialogTitle, setPasswordDialogTitle] = React.useState<string>('');
  const [password, setPassword] = React.useState<string|null>(null);
  const userService = useUserService();

  const resetAddDialog = () => {
    setEmail('');
    setName('');
    setIsEmailValid(false);
    setCreating(false);
    setError(null);
  };

  const openCreationDialog = () => {
    resetAddDialog();
    setCreationDialogOpened(true);
  };

  const closeCreationDialog = () => {
    resetAddDialog();
    setCreationDialogOpened(false);
  };

  const closePasswordDialog = () => {
    setPasswordDialogOpened(false);
    setPasswordDialogTitle('');
    setPassword('');
  };

  const createUser = () => {
    setError(null);
    setCreating(true);
    userService.create(email, name, (resultText, user) => {
      setCreating(false);
      if (resultText == 'success' && user != null) {
        props.onCreation(user);
        setCreationDialogOpened(false);
        resetAddDialog();
        if (user.password) {
          setPassword(user.password);
          setPasswordDialogTitle(`Password for ${user.email}`);
          setPasswordDialogOpened(true);
        }
      } else {
        setError(`Failed to create user (${resultText})`);
      }
    });
  };

  return (
    <Box>
      <IconButton onClick={openCreationDialog} sx={props.sx} ><PersonAddRounded /></IconButton>
      {/* Creation Form */}
      <Dialog open={creationDialogOpened} >
        <DialogTitle>Add a new user</DialogTitle>
        { error != null && (<Alert severity="error">{error}</Alert>) }
        <DialogContent>
          <Typography>Create a new user</Typography>
          <TextField
            fullWidth
            required
            error={email.length > 0 && !isEmailValid}
            onChange={(event) => {
              const input = event.target.value;
              const valid = validator.isEmail(input);
              setEmail(input);
              setIsEmailValid(valid);
            }}
            margin="dense"
            id="email"
            label="Email"
          />
          <TextField
            fullWidth
            margin="dense"
            id="name"
            label="Name"
            onChange={(event) => {
              const input = event.target.value;
              setName(input);
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={closeCreationDialog}>Cancel</Button>
          <Button disabled={!isEmailValid || creating} onClick={() => createUser()}>Add</Button>

        </DialogActions>
      </Dialog>
      {/* New User Password */}
      <PasswordDialog open={passwordDialogOpened} title={passwordDialogTitle} password={password} onClose={closePasswordDialog} />
    </Box>
  );
}

export default UserAdd;
