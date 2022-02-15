import React from 'react';
import {Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton} from '@mui/material';
import {DeleteRounded} from '@mui/icons-material';
import {Trans, useTranslation} from 'react-i18next';
import {User} from '../../../../models/user';
import {useUserService} from '../../../../services/user/user-service';

function UserDelete(props:{user: User, onDelete:(user: User) => void}) {
  const [confirmationOpened, setConfirmationOpened] = React.useState<boolean>(false);
  const userService = useUserService();
  const {t} = useTranslation();

  const deleteButtonPressed = () => {
    setConfirmationOpened(true);
  };

  const cancelUserDelete = () => {
    setConfirmationOpened(false);
  };

  const confirmUserDelete = () => {
    setConfirmationOpened(false);
    userService.delete(props.user.email, (resultText) => {
      if (resultText == 'success') {
        props.onDelete(props.user);
      }
    });
  };

  return (
    <Box>
      <IconButton aria-label={t('userDelete.ariaLabel', {email: props.user.email})} onClick={deleteButtonPressed}>
        <DeleteRounded />
      </IconButton>
      {/* Confirm Delete */}
      <Dialog open={confirmationOpened}>
        <DialogTitle><Trans i18nKey='userDelete.dialogTitle' values={{email: props.user.email}} /></DialogTitle>
        <DialogContent><Trans i18nKey='userDelete.dialogInstruction' /></DialogContent>
        <DialogActions>
          <Button onClick={cancelUserDelete}><Trans i18nKey='actions.cancel' /></Button>
          <Button onClick={confirmUserDelete} color="error"><Trans i18nKey='actions.delete'/></Button>
        </DialogActions>
      </Dialog>
    </Box>);
}

export default UserDelete;
