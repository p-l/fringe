import React, {ErrorInfo} from 'react';
import {Alert} from '@mui/material';

interface State {
  hasError: boolean;
}

const ErrorMessage: React.FC = () => {
  return (
    <Alert severity="error">The application could not be loaded properly.</Alert>
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
    console.log('error:'+error.message);
  }

  render() {
    if (this.state.hasError) {
      return <ErrorMessage />;
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
