import React from 'react';
import {Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, Snackbar} from '@mui/material';
import {DeleteRounded} from '@mui/icons-material';
import {User} from '../../../../models/user';
import {useUserService} from '../../../../services/user/user-service';

function UserDelete(props:{user: User, onDelete:(user: User) => void}) {
  const [confirmationOpened, setConfirmationOpened] = React.useState<boolean>(false);
  const [deleted, setDeleted] = React.useState<boolean>(false);
  const userService = useUserService();

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
        setDeleted(true);
        props.onDelete(props.user);
      }
    });
  };

  return (
    <Box>
      <IconButton onClick={deleteButtonPressed}>
        <DeleteRounded />
      </IconButton>
      {/* Confirm Delete */}
      <Dialog open={confirmationOpened}>
        <DialogTitle>Delete {props.user.email}?</DialogTitle>
        <DialogContent>Remove user from the database. This does not prevent the user from enrolling again later.</DialogContent>
        <DialogActions>
          <Button onClick={cancelUserDelete}>Cancel</Button>
          <Button onClick={confirmUserDelete} color="error">Delete</Button>
        </DialogActions>
      </Dialog>
      <Snackbar
        autoHideDuration={3000}
        open={deleted}
        onClose={() => setDeleted(false)}
        message={`User ${props.user.email} was deleted`}
      />
    </Box>);
}

export default UserDelete;
