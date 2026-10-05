import{j as t}from"./jsx-runtime-BTJTZTIL.js";import{F as a,e as d,g as u}from"./index.es-C0EmWI2V.js";import{w as C,D as k}from"./index-DQqI27CC.js";import{w as S,e as D,f as N}from"./index-Ch1O_M5-.js";import{B as o,T as s}from"./Button-DTtPuz4n.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-2AqrBsEA.js";import"./iframe-CjolCmMi.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-DgH-xKnr.js";import"./index-DrFu-skq.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";const P=()=>{const e=C();return t.jsxs("div",{className:"button-docs",children:[t.jsx("style",{children:`
                .button-docs .sb-anchor > h3:first-child {
                    border-bottom: 1px solid ${e.appBorderColor};
                    font-size: 24px;
                    padding-bottom: 4px;
                }
            `}),t.jsx(k,{})]})},K={title:"Components/Button",component:o,parameters:{layout:"centered",docs:{page:P,description:{component:`
### Use a Button to perform an action

Use a [Link](http://localhost:6006/?path=/docs/components-link--docs) to navigate to another location.

Use a Button when a user remains in the current context and triggers an action, such as:

- Submitting or resetting a form
- Saving, creating, deleting, or updating data
- Opening a dialog, menu, popover, or drawer
- Expanding or collapsing content
- Starting, stopping, or retrying a process
- Moving through a multi-step interface without changing the URL.

### Choosing a variant

- Use primary for the main action in a section or workflow.
- Use secondary for supporting actions such as Cancel or Back.
- Use TextButton for lower-emphasis text actions.
- Use IconButton for icon-only actions.

### Accessibility

- Give every button a clear accessible name describing its action.
- Use visible text when possible. Icon-only buttons require an aria-label.
- Preserve the visible keyboard focus indicator.
- Do not nest a Button inside a Link or another interactive element.
- Use disabled only when the action is unavailable. Explain why when the reason is not apparent.

### Forms and rendered elements

Button defaults to \`type="button"\`, preventing it from unintentionally submitting a containing form. Pass \`type="submit"\` when the Button should submit the form:

\`\`\`tsx
<form onSubmit={handleSubmit}>
    <Button>Cancel</Button>
    <Button type='submit'>Save</Button>
</form>
\`\`\`

The Cancel Button renders with \`type="button"\`, while the Save Button keeps the supplied \`type="submit"\`. You can also pass \`type="reset"\` when native form reset behavior is intended.

Use the Base UI \`render\` prop when another element needs to provide the rendered structure. This replaces the former \`asChild\` pattern and avoids nesting interactive elements.

### Deprecated APIs

The \`size\` and \`fontColor\` props and the transparent and icon variants are legacy APIs. Prefer the dedicated \`TextButton\` and \`IconButton\` components for new code.

### Avoid choosing by appearance

Buttons and Links may be styled similarly, but their semantics should reflect their behavior:

- Do not use a Link solely because the control looks like text.
- Do not use a Button solely because the control looks prominent.
- Do not place a Button inside a Link or a Link inside a Button.

Using the correct element provides expected keyboard behavior and helps assistive technology communicate what will happen.
                `}}},tags:["autodocs"],argTypes:{render:{description:"Use as a replacement for the `asChild` prop"},variant:{options:["primary","secondary"],control:"select"},fontColor:{description:"**Deprecated:** Use `TextButton` instead. This prop will be removed in BED-7642.",control:!1,table:{type:{summary:"'primary' | null"}}},size:{options:["small","medium","large"],control:"select",description:"deprecated",table:{category:"Deprecated"}},disabled:{control:"boolean",description:"Disables button interactions."}},args:{onClick:N()}},r={args:{children:"Type is Button",disabled:!1,variant:"primary",size:"medium"},argTypes:{type:{description:"Sets the native HTML button type.",options:["button","submit","reset"],control:{type:"select"},table:{category:"HTML attributes",defaultValue:{summary:"button"},type:{summary:"'button' | 'submit' | 'reset'"}}}},parameters:{docs:{description:{story:'Button defaults to `type="button"` so it does not unintentionally submit a containing form. Use the type control to verify that an explicitly supplied `type="submit"` or `type="reset"` overrides the default.'}}},render:({children:e,...n})=>t.jsx(o,{...n,children:e}),play:async({canvasElement:e,args:n})=>{const U=await S(e).findByRole("button");D(U).toHaveAttribute("type",n.type??"button")}},i={args:{variant:"primary",children:"Next",disabled:!1},parameters:{docs:{description:{story:"Use the primary variant for the main action in a section or workflow. Prefer one primary action per decision area."}}},render:({children:e,...n})=>t.jsxs(t.Fragment,{children:[t.jsx("div",{className:"flex justify-center mb-10",children:t.jsx(o,{...n,children:e})}),t.jsx("hr",{className:"mb-10"}),t.jsxs("div",{className:"flex items-center gap-4",children:[t.jsxs(o,{children:[e,t.jsx(a,{icon:d})]}),t.jsx(o,{variant:"primary",disabled:!0,children:"Disabled"}),t.jsxs(o,{variant:"primary",disabled:!0,children:[t.jsx(a,{icon:d}),"Disabled"]})]})]})},c={args:{variant:"secondary",children:"Secondary",disabled:!1},parameters:{docs:{description:{story:"Use the secondary variant for supporting actions such as Cancel, Back, or an alternative to the primary action."}}},render:({variant:e,children:n,...p})=>t.jsxs(t.Fragment,{children:[t.jsx("div",{className:"flex justify-center mb-10",children:t.jsx(o,{variant:e,...p,children:n})}),t.jsx("hr",{className:"mb-10"}),t.jsxs("div",{className:"flex items-center gap-4",children:[t.jsxs(o,{variant:"secondary",children:[n,t.jsx(a,{icon:d})]}),t.jsx(o,{variant:"secondary",disabled:!0,children:"Disabled"}),t.jsxs(o,{variant:"secondary",disabled:!0,children:[t.jsx(a,{icon:d}),"Disabled"]})]})]})},l={args:{children:"Help Text",disabled:!1,fontColor:"default"},argTypes:{fontColor:{control:"select",options:["primary","default"]}},parameters:{docs:{description:{story:`### Usage

Use TextButton for lower-emphasis actions that keep the user in the current context. It is still a button, not a navigation link.

- Use \`fontColor="primary"\` when the action needs additional emphasis.
- Use \`disabled\` when the action is unavailable.
- Use \`IconButton\` instead when an action has no visible text label.
- Use the \`render\` prop when integrating with another component without nesting interactive elements.
- Use \`className\` for local layout adjustments such as padding, alignment, or width; preserve the standard focus indicator.

~~~tsx
<TextButton fontColor='primary' onClick={clearFilters}>
    Clear All
</TextButton>
~~~`}}},render:({children:e,...n})=>t.jsxs(t.Fragment,{children:[t.jsx("div",{className:"flex justify-center mb-10",children:t.jsx(s,{...n,children:e})}),t.jsx("hr",{className:"mb-10"}),t.jsxs("div",{className:"flex items-center gap-4",children:[t.jsxs(s,{children:[e,t.jsx(a,{icon:u})]}),t.jsx(s,{fontColor:"primary",children:"Primary Text"}),t.jsxs(s,{disabled:!0,children:[t.jsx(a,{icon:u}),"Disabled"]})]})]})};var m,h,y;r.parameters={...r.parameters,docs:{...(m=r.parameters)==null?void 0:m.docs,source:{originalSource:`{
  args: {
    children: 'Type is Button',
    disabled: false,
    variant: 'primary',
    size: 'medium'
  },
  argTypes: {
    type: {
      description: 'Sets the native HTML button type.',
      options: ['button', 'submit', 'reset'],
      control: {
        type: 'select'
      },
      table: {
        category: 'HTML attributes',
        defaultValue: {
          summary: 'button'
        },
        type: {
          summary: "'button' | 'submit' | 'reset'"
        }
      }
    }
  },
  parameters: {
    docs: {
      description: {
        story: \`Button defaults to \\\`type="button"\\\` so it does not unintentionally submit a containing form. Use the type control to verify that an explicitly supplied \\\`type="submit"\\\` or \\\`type="reset"\\\` overrides the default.\`
      }
    }
  },
  render: ({
    children,
    ...buttonProps
  }) => {
    return <Button {...buttonProps}>{children}</Button>;
  },
  play: async ({
    canvasElement,
    args
  }) => {
    const canvas = within(canvasElement);
    const button = await canvas.findByRole('button');

    // Button defaults to \`type="button"\` so it does not
    // unintentionally submit a containing form.
    expect(button).toHaveAttribute('type', args.type ?? 'button');
  }
}`,...(y=(h=r.parameters)==null?void 0:h.docs)==null?void 0:y.source}}};var b,f,x;i.parameters={...i.parameters,docs:{...(b=i.parameters)==null?void 0:b.docs,source:{originalSource:`{
  args: {
    variant: 'primary',
    children: 'Next',
    disabled: false
  },
  parameters: {
    docs: {
      description: {
        story: \`Use the primary variant for the main action in a section or workflow. Prefer one primary action per decision area.\`
      }
    }
  },
  render: ({
    children,
    ...buttonProps
  }) => <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <Button {...buttonProps}>{children}</Button>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-4'>
                <Button>
                    {children}
                    <FontAwesomeIcon icon={faStar} />
                </Button>
                <Button variant='primary' disabled>
                    Disabled
                </Button>
                <Button variant='primary' disabled>
                    <FontAwesomeIcon icon={faStar} />
                    Disabled
                </Button>
            </div>
        </>
}`,...(x=(f=i.parameters)==null?void 0:f.docs)==null?void 0:x.source}}};var v,g,B;c.parameters={...c.parameters,docs:{...(v=c.parameters)==null?void 0:v.docs,source:{originalSource:`{
  args: {
    variant: 'secondary',
    children: 'Secondary',
    disabled: false
  },
  parameters: {
    docs: {
      description: {
        story: \`Use the secondary variant for supporting actions such as Cancel, Back, or an alternative to the primary action.\`
      }
    }
  },
  render: ({
    variant,
    children,
    ...buttonProps
  }) => <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <Button variant={variant} {...buttonProps}>
                    {children}
                </Button>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-4'>
                <Button variant='secondary'>
                    {children}
                    <FontAwesomeIcon icon={faStar} />
                </Button>
                <Button variant='secondary' disabled>
                    Disabled
                </Button>
                <Button variant='secondary' disabled>
                    <FontAwesomeIcon icon={faStar} />
                    Disabled
                </Button>
            </div>
        </>
}`,...(B=(g=c.parameters)==null?void 0:g.docs)==null?void 0:B.source}}};var w,T,j;l.parameters={...l.parameters,docs:{...(w=l.parameters)==null?void 0:w.docs,source:{originalSource:`{
  args: {
    children: 'Help Text',
    disabled: false,
    fontColor: 'default'
  },
  argTypes: {
    fontColor: {
      control: 'select',
      options: ['primary', 'default']
    }
  },
  parameters: {
    docs: {
      description: {
        story: \`### Usage

Use TextButton for lower-emphasis actions that keep the user in the current context. It is still a button, not a navigation link.

- Use \\\`fontColor="primary"\\\` when the action needs additional emphasis.
- Use \\\`disabled\\\` when the action is unavailable.
- Use \\\`IconButton\\\` instead when an action has no visible text label.
- Use the \\\`render\\\` prop when integrating with another component without nesting interactive elements.
- Use \\\`className\\\` for local layout adjustments such as padding, alignment, or width; preserve the standard focus indicator.

~~~tsx
<TextButton fontColor='primary' onClick={clearFilters}>
    Clear All
</TextButton>
~~~\`
      }
    }
  },
  render: ({
    children,
    ...textButtonProps
  }) => <>
            {/* Storybook controls affect only this button */}
            <div className='flex justify-center mb-10'>
                <TextButtonComponent {...textButtonProps}>{children}</TextButtonComponent>
            </div>
            <hr className='mb-10' />
            {/* These buttons remain static */}
            <div className='flex items-center gap-4'>
                <TextButtonComponent>
                    {children}
                    <FontAwesomeIcon icon={faListUl} />
                </TextButtonComponent>
                <TextButtonComponent fontColor='primary'>Primary Text</TextButtonComponent>
                <TextButtonComponent disabled>
                    <FontAwesomeIcon icon={faListUl} />
                    Disabled
                </TextButtonComponent>
            </div>
        </>
}`,...(j=(T=l.parameters)==null?void 0:T.docs)==null?void 0:j.source}}};const Q=["DefaultType","Primary","Secondary","TextButton"];export{r as DefaultType,i as Primary,c as Secondary,l as TextButton,Q as __namedExportsOrder,K as default};
