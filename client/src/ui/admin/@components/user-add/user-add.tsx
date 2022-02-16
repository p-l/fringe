import {PersonAddRounded} from '@mui/icons-material';
import {Alert, Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, SxProps, TextField, Theme, Typography} from '@mui/material';
import React from 'react';
import {Trans, useTranslation} from 'react-i18next';
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
  const {t} = useTranslation();

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
          setPasswordDialogTitle(t('passwordDialog.newUserTitle', {email: user.email}));
          setPasswordDialogOpened(true);
        }
      } else {
        setError(t('userAdd.failure', {resultCode: resultText}));
      }
    });
  };

  return (
    <Box>
      <IconButton aria-label={t('userAdd.ariaLabel')} onClick={openCreationDialog} sx={props.sx} >
        <PersonAddRounded />
      </IconButton>
      {/* Creation Form */}
      <Dialog open={creationDialogOpened} >
        <DialogTitle><Trans i18nKey='userAdd.dialogTitle' /></DialogTitle>
        { error != null && (<Alert severity="error">{error}</Alert>) }
        <DialogContent>
          <Typography><Trans i18nKey='userAdd.dialogInstruction' /></Typography>
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
            label={t('userAdd.emailLabel')}
          />
          <TextField
            fullWidth
            margin="dense"
            id="name"
            label={t('userAdd.nameLabel')}
            onChange={(event) => {
              const input = event.target.value;
              setName(input);
            }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={closeCreationDialog}><Trans i18nKey='actions.cancel' /></Button>
          <Button disabled={!isEmailValid || creating} onClick={() => createUser()}><Trans i18nKey='actions.add' /></Button>

        </DialogActions>
      </Dialog>
      {/* New User Password */}
      <PasswordDialog open={passwordDialogOpened} title={passwordDialogTitle} password={password} onClose={closePasswordDialog} />
    </Box>
  );
}

export default UserAdd;
