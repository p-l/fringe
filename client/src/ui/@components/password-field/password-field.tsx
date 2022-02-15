import React from 'react';
import {Alert, Box, CircularProgress, IconButton, InputAdornment, Snackbar, SxProps, TextField, Theme} from '@mui/material';
import {CheckCircleOutlineRounded, CopyAllRounded} from '@mui/icons-material';
import {Trans, useTranslation} from 'react-i18next';

function PasswordField(props: {loading: boolean, password?: string|null|undefined, sx?: SxProps<Theme>|undefined }) {
  const [passwordCopied, setPasswordCopied] = React.useState<boolean>(false);
  const {t} = useTranslation();

  const copyUserPassword = () => {
    if (props.password != null) {
      navigator.clipboard.writeText(props.password).then(() => {
        setPasswordCopied(true);
      });
    }
  };

  return (
    <Box sx={props.sx}>
      { /* Password Copy Confirmation */ }
      <Snackbar
        anchorOrigin={{vertical: 'bottom', horizontal: 'center'}}
        autoHideDuration={3000}
        open={passwordCopied}
        onClose={() => setPasswordCopied(false)}
      >
        <Alert variant="filled" severity="success" sx={{width: '100%'}}><Trans i18nKey='passwordField.copied'/></Alert>
      </Snackbar>
      <TextField
        fullWidth
        margin="dense"
        id="fringe-user-password"
        label={t('passwordField.label')}
        variant="standard"
        disabled={true}
        type="text"
        value={props.loading ? 'Loading' : props.password }
        InputProps={{
          startAdornment: props.loading ? <InputAdornment position="start"><CircularProgress size={14} /></InputAdornment> : '',
          endAdornment: props.loading ? '' : <InputAdornment position="end"><IconButton aria-label={t('passwordField.copy')} onClick={copyUserPassword} edge="end">{passwordCopied ? <CheckCircleOutlineRounded /> : <CopyAllRounded />}</IconButton></InputAdornment>,
        }}
      />
    </Box>
  );
}

export default PasswordField;
