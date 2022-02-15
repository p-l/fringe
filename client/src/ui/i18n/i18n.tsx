import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

i18n
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
      debug: true,
      fallbackLng: 'en',
      interpolation: {
        escapeValue: false,
      },
      resources: {
        en: {
          translation: {
            actions: {
              add: 'Add',
              cancel: 'Cancel',
              close: 'Close',
              delete: 'Delete',
              reset: 'Reset',
              understood: 'Understood',
            },
            admin: {
              headerEmail: 'Email',
              headerName: 'Name',
              headerLastSeen: 'Last Seen',
              headerPasswordAge: 'Password Change',
              lastSeen: '{{lastSeenDate, datetime}}',
              loadMoreUsers: 'More Users',
              noMoreUsers: 'No More User',
              passwordAge_zero: 'Today',
              passwordAge_other: '{{passwordAge, relativetime(day)}}',
            },
            errorBoundary: {
              userMessage: 'The application could not be loaded properly.',
            },
            login: {
              errorFringeRejected: 'Authentication rejected by Fringe server.',
              errorGoogleFailed: 'Google refused authentication or provided an invalid response',
              signInButton: 'Sign-in With Google',
              welcome: 'Welcome to Fringe',
            },
            me: {
              errorFailedToGetPassword: 'Could not retrieve new password from server',
              lastSeen: 'Last Authentication',
              lastSeenDate: '{{lastSeenDate, datetime}}',
              newPassword: 'New Password',
              passwordAge: 'Last Password Change',
              passwordAgeRelative_zero: 'Today',
              passwordAgeRelative_other: '{{passwordAge, relativetime(day)}}',
              renewDialogTitle: 'Replace Current Password',
              renewDialogInstruction: 'Replace your current password with a new password will remove the current password. Only the new password will be accepted\'',
            },
            passwordDialog: {
              newUserTitle: 'Password for {{email}}',
              existingUserTitle: 'New password for {{email}}',
              instructions: 'Make sure to copy the password below. Once the dialog closed it cannot be retrieved again.',
            },
            passwordField: {
              label: 'Password',
              copied: 'Password was copied to clipboard',
              copy: 'copy password to clipboard',
            },
            userAdd: {
              ariaLabel: 'Add user',
              dialogTitle: 'Add a new user',
              dialogInstruction: 'Create a new user',
              emailLabel: 'Email',
              failure: 'Failed to create user ({{resultCode}})',
              nameLabel: 'Full Name',
            },
            userDelete: {
              ariaLabel: 'Delete user {{email}}',
              dialogTitle: 'Delete {{email}}?',
              dialogInstruction: 'Remove user from the database. This does not prevent the user from enrolling again later.',
            },
            userRenew: {
              ariaLabel: 'New password for {{email}}',
              dialogTitle: 'Reset {{email}}\'s password?',
              dialogInstruction: 'Replaces {{email}}\'s password with a new password. Only the new password will be accepted',
            },
            userSearch: {
              clearButtonAriaLabel: 'Clear search',
              fieldAriaLabel: 'User search',
              searchButtonAriaLabel: 'Do Search',
            },
          },
        },
      },
    });

export default i18n;
