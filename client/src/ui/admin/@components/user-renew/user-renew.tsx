import React from 'react';
import {Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton} from '@mui/material';
import {PasswordRounded} from '@mui/icons-material';
import {Trans, useTranslation} from 'react-i18next';

import {User} from '../../../../models/user';
import {useUserService} from '../../../../services/user/user-service';
import PasswordDialog from '../password-dialog';


function UserRenew(props:{user: User}) {
  const [showConfirmation, setShowConfirmation] = React.useState<boolean>(false);
  const [showPassword, setShowPassword] = React.useState<boolean>(false);
  const [password, setPassword] = React.useState<string|null>( null);
  const userService = useUserService();
  const {t} = useTranslation();

  const resetUserPassword = () => {
    setShowConfirmation(true);
    setShowPassword(false);
    setPassword(null);
  };

  const confirmUserReset = () => {
    setShowConfirmation(false);
    setShowPassword(true);
    userService.renewPassword(props.user.email, (user) => {
      if (user != null && user.password != null && user.password.length > 0) {
        setPassword(user.password);
      } else {
        console.warn(`Failed to get user's password`);
      }
    });
  };

  const cancelUserReset = () => {
    setShowConfirmation(false);
    setShowPassword(false);
    setPassword(null);
  };

  const closePasswordDialog = () => {
    setShowConfirmation(false);
    setShowPassword(false);
    setPassword(null);
  };


  return (
    <Box>
      <IconButton aria-label={t('userRenew.ariaLabel', {email: props.user.email})} onClick={() => resetUserPassword()}>
        <PasswordRounded />
      </IconButton>
      {/* Confirm Password Reset */}
      <Dialog open={showConfirmation}>
        <DialogTitle><Trans i18nKey='userRenew.dialogTitle' values={{email: props.user.email}} /></DialogTitle>
        <DialogContent><Trans i18nKey='userRenew.dialogInstruction' values={{email: props.user.email}} /></DialogContent>
        <DialogActions>
          <Button onClick={cancelUserReset}><Trans i18nKey='actions.cancel' /></Button>
          <Button onClick={confirmUserReset} color="error"><Trans i18nKey='actions.reset' /></Button>
        </DialogActions>
      </Dialog>
      {/* Show new Password */}
      <PasswordDialog open={showPassword} title={t('passwordDialog.existingUserTitle', {email: props.user.email})} password={password} onClose={closePasswordDialog} />
    </Box>
  );
}

export default UserRenew;
