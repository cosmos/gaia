import React from 'react';

export const Var = ({ children }) => {
  return <code>{JSON.stringify(children)}</code>;
};
