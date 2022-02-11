import React from 'react';
import {Box, CircularProgress, FormControl, IconButton, Input, InputAdornment, Snackbar, SxProps, Theme} from '@mui/material';
import {CheckCircleOutlineRounded, CopyAllRounded} from '@mui/icons-material';

function PasswordField(props: {loading: boolean, password?: string|null|undefined, sx?: SxProps<Theme>|undefined }) {
  const [passwordCopied, setPasswordCopied] = React.useState<boolean>(false);

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
        autoHideDuration={3000}
        open={passwordCopied}
        onClose={() => setPasswordCopied(false)}
        message="Password was copied to clipboard"
      />
      <FormControl variant="standard" sx={{width: 1}}>
        {/* <InputLabel htmlFor="fringe-user-password">Password</InputLabel>*/}
        <Input
          id="fringe-user-password"
          readOnly={true}
          disabled={true}
          type="text"
          value={props.loading ? 'Loading' : props.password }
          startAdornment={
            props.loading && (<InputAdornment position="start"><CircularProgress size={14} /></InputAdornment>)
          }
          endAdornment={!props.loading && (
            <InputAdornment position="end">
              <IconButton
                aria-label="copy password to clipboard"
                onClick={copyUserPassword}
                disabled={props.loading}
                edge="end"
              >
                { passwordCopied && (<CheckCircleOutlineRounded />) }
                { !passwordCopied && (<CopyAllRounded />) }
              </IconButton>
            </InputAdornment>
          )
          }
        />
      </FormControl>
    </Box>
  );
}

export default PasswordField;
