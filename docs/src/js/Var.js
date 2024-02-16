import React from 'react';

export const Var = ({ children }) => {
  return <code>{JSON.stringify(children)}</code>;
};

export const PlainVar = ({ children }) => {
  return <span>{JSON.stringify(children)}</span>;
};

export const ConsoleOutput = ({ children }) => {
  
  return (
    <textarea readonly className='text-left'>
      {JSON.stringify(children.replace(/\n/g, "\r"))}
    </textarea>
  )
};
