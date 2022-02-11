import {ClearRounded, PersonSearchRounded} from '@mui/icons-material';
import {IconButton, Input, InputAdornment, SxProps, Theme} from '@mui/material';
import React from 'react';

let searchTimer = setTimeout(()=>{}, 0);

function UserSearch(props: {onSearch : (query: string) => void, onClear: () => void, delay?: number|undefined, sx?: SxProps<Theme>|undefined }) {
  const [search, setSearch] = React.useState<string>('');
  const delay = props.delay ? props.delay : 750;

  const doSearch = (searchString: string, delayed: boolean = true) => {
    setSearch(searchString);
    clearTimeout(searchTimer);

    if (searchString != '') {
      searchTimer = setTimeout(() => {
        props.onSearch(searchString);
      }, delayed ? delay : 0);
    } else {
      props.onClear();
    }
  };

  return (
    <Input
      id='user-search'
      sx={props.sx}
      value={search}
      onChange={(event) => doSearch(event.target.value)}
      onKeyPress={(event) => {
        if (event.key === 'Enter') {
          const thisInput = document.getElementById('user-search') as HTMLInputElement;
          doSearch(thisInput.value, false);
        }
      }}
      startAdornment={<InputAdornment position='start'><IconButton onClick={() => doSearch(search, false)}><PersonSearchRounded /></IconButton></InputAdornment> }
      endAdornment={(search.length > 0) && <InputAdornment position='end'><IconButton onClick={ () => doSearch('', false)}><ClearRounded /></IconButton></InputAdornment>}
    />
  );
}

export default UserSearch;
