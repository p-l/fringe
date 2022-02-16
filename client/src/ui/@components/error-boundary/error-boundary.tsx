import React, {ErrorInfo} from 'react';
import {Alert} from '@mui/material';
import {Trans} from 'react-i18next';

interface State {
  hasError: boolean;
}

const ErrorMessage: React.FC = () => {
  return (

    <Alert severity="error">
      <Trans i18nKey="errorBoundary.userMessage" />
    </Alert>
  );
};

export class ErrorBoundary extends React.Component<{}, State> {
  static getDerivedStateFromError() {
    return {hasError: true};
  }

  state: State = {
    hasError: false,
  };

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.debug(`ErrorBoundary did catch: ${error.message}`);
  }

  render() {
    if (this.state.hasError) {
      return <ErrorMessage />;
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
