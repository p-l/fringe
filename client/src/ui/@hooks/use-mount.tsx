import {useEffect} from 'react';

const useMountEffect = (fun:VoidFunction) => useEffect(fun, []);

export default useMountEffect;
