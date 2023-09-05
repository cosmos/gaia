import React from 'react';

const defaultEntryRenderer = (key, value) => (
  <tr key={`${key}-${value}`}>
    <td>
      <code>{key}</code>
    </td>
    <td>
      <code>{JSON.stringify(value)}</code>
    </td>
  </tr>
);

export const Table = ({
  data = {},
  renderEntry = defaultEntryRenderer,
  ...otherProps
}) => {
  return (
    <table {...otherProps}>
      <tr className="text-left">
        <th className="px-5.5">Key</th>
        <th className="px-5.5">Value</th>
      </tr>

      {Object.entries(data).map(([key, value]) => renderEntry(key, value))}
    </table>
  );
};
