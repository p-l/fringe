import {Button, Dialog, DialogActions, DialogContent, DialogTitle} from '@mui/material';
import React from 'react';
import PasswordField from '../../../@components/password-field';

function PasswordDialog(props:{open: boolean, title: string, password: string|null, onClose: VoidFunction}) {
  return (
    <Dialog open={props.open}>
      <DialogTitle>{props.title}</DialogTitle>
      <DialogContent>
                Make sure to copy the password below. Once the dialog closed it cannot be retrieved again.
      </DialogContent>
      <DialogContent sx={{topMargin: 0}}>
        <PasswordField loading={props.password == null || props.password.length == 0} password={props.password} sx={{width: 1}} />
      </DialogContent>
      <DialogActions>
        <Button onClick={props.onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}

export default PasswordDialog;
