import{j as r}from"./jsx-runtime-BTJTZTIL.js";import{L as x}from"./Link-Q3xt4pq6.js";import"./index-ChsGqxH_.js";import"./index.es-C0EmWI2V.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-2AqrBsEA.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";const v={title:"Components/Link",component:x,tags:["autodocs"],args:{href:"https://example.com",children:"External link",variant:"styled"},argTypes:{variant:{control:"select",options:["styled","unstyled"]}},parameters:{docs:{description:{component:`
### When to use:

- **Use this for external links**

This component:
- Always opens links in a new tab
- Always applies \`noopener\`
- Applies \`noreferrer\` by default
- Supports \`allowReferrer\` to omit \`noreferrer\`
- Displays an external link icon
- Announces "(opens in a new tab)" to screen readers
        `}}}},e={},t={args:{variant:"unstyled",children:"Unstyled external link"}},n={name:"In text",render:u=>r.jsxs("p",{className:"text-sm text-neutral-dark-1 dark:text-white",children:["Read the ",r.jsx(x,{...u})," for more details."]}),args:{href:"https://example.com",children:"documentation",variant:"styled"},parameters:{docs:{description:{story:`
  ### Inline usage
  
  **Do** use explicit spacing:
  
  \`\`\`tsx
  Read the{' '}
  <Link href="https://example.com">documentation</Link>{' '}
  for more details
  \`\`\`
  
  **Don’t** rely on implicit spacing.
          `},source:{code:`<p className='text-sm text-neutral-dark-1 dark:text-white'>
    Read the{' '}
    <Link href='https://example.com'>documentation</Link>{' '}
    for more details.
  </p>`}}}};var a,s,o;e.parameters={...e.parameters,docs:{...(a=e.parameters)==null?void 0:a.docs,source:{originalSource:"{}",...(o=(s=e.parameters)==null?void 0:s.docs)==null?void 0:o.source}}};var i,l,p;t.parameters={...t.parameters,docs:{...(i=t.parameters)==null?void 0:i.docs,source:{originalSource:`{
  args: {
    variant: 'unstyled',
    children: 'Unstyled external link'
  }
}`,...(p=(l=t.parameters)==null?void 0:l.docs)==null?void 0:p.source}}};var c,m,d;n.parameters={...n.parameters,docs:{...(c=n.parameters)==null?void 0:c.docs,source:{originalSource:`{
  name: 'In text',
  render: args => <p className='text-sm text-neutral-dark-1 dark:text-white'>
            Read the <Link {...args} /> for more details.
        </p>,
  args: {
    href: 'https://example.com',
    children: 'documentation',
    variant: 'styled'
  },
  parameters: {
    docs: {
      description: {
        story: \`
  ### Inline usage
  
  **Do** use explicit spacing:
  
  \\\`\\\`\\\`tsx
  Read the{' '}
  <Link href="https://example.com">documentation</Link>{' '}
  for more details
  \\\`\\\`\\\`
  
  **Don’t** rely on implicit spacing.
          \`
      },
      source: {
        code: \`<p className='text-sm text-neutral-dark-1 dark:text-white'>
    Read the{' '}
    <Link href='https://example.com'>documentation</Link>{' '}
    for more details.
  </p>\`
      }
    }
  }
}`,...(d=(m=n.parameters)==null?void 0:m.docs)==null?void 0:d.source}}};const I=["Styled","Unstyled","InText"];export{n as InText,e as Styled,t as Unstyled,I as __namedExportsOrder,v as default};
