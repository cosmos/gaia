import React from 'react';

export const KeyValueTable = ({
  data = {},
  renderKey = (key) => (
    <a href={`#${key.toLowerCase()}`}>
      <code>{key}</code>
    </a>
  ),
  renderValue = (value) => value,
  ...otherProps
}) => {
  return (
    <table {...otherProps}>
      <tr className="text-left">
        <th className="px-5.5">Key</th>
        <th className="px-5.5">Value</th>
      </tr>

      {Object.entries(data).map(([key, value]) => (
        <tr key={`${key}-${value}`}>
          <td>{renderKey(key)}</td>
          <td>
            <code>{JSON.stringify(renderValue(value))}</code>
          </td>
        </tr>
      ))}
    </table>
  );
};
