import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{D as l,a as s,b as c,c as d,d as g,e as u,f as h,g as p}from"./Dialog-dymNnUUK.js";import{B as o}from"./Button-DTtPuz4n.js";import{V as j}from"./VisuallyHidden-DT_8ZbEK.js";import"./index-ChsGqxH_.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-Dv0R_RYs.js";import"./index-Dc_EP5I0.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-ejpvWIAw.js";import"./index-BWEBPL6D.js";import"./index-DnVAETpg.js";import"./component-D-Cd_cec.js";import"./index-BQHA1h7N.js";import"./index-vQUXxau3.js";import"./Combination-HODMvkiJ.js";import"./index-Bb4qSo10.js";import"./Typography-BMbK6Q6M.js";import"./utils-BIldDEfN.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";import"./index-oIeaYTeP.js";const Y={title:"Components/Dialog",component:l,parameters:{layout:"centered",docs:{description:{component:"A modal dialog that overlays the page to focus the user on a single, self-contained task. Use it to confirm destructive actions, request input that interrupts the current flow, or surface contextual information that the user must acknowledge before continuing."}}},tags:["autodocs"],argTypes:{DialogOverlayProps:{control:"select",options:["default","blur"],mapping:{blur:{blurBackground:!0}},description:"Props forwarded to the underlying `DialogOverlay`. Use the `blur` option to enable a backdrop blur (`blurBackground: true`) behind the dialog.",table:{type:{summary:"DialogOverlayProps"},defaultValue:{summary:"{ blurBackground: false }"}}},disableStickyLayout:{control:"boolean",description:"When `true`, opts out of the default sticky layout. By default, `DialogTitle` and `DialogActions` are pinned to the top and bottom of `DialogContent` and any other children are placed inside an internal scrollable container, so the title and actions remain in view while the middle content scrolls. With this prop enabled, all children render in source order and the entire dialog scrolls as a single block — a difference that only becomes visible when the content overflows.",table:{type:{summary:"boolean"},defaultValue:{summary:"false"}}},maxWidth:{control:"select",options:["xl","lg","md","sm","xs"],description:"Constrains the maximum width of the dialog. Values match the MUI breakpoint scale: `xs` (444px), `sm` (600px), `md` (900px), `lg` (1200px), `xl` (1536px).",table:{type:{summary:"'xl' | 'lg' | 'md' | 'sm' | 'xs'"},defaultValue:{summary:"sm"}}},allowNav:{control:"boolean",description:"When `true`, lowers the overlay z-index and re-enables pointer events on the document body so the user can still interact with the surrounding application (for example, switching tabs) while the dialog is open.",table:{type:{summary:"boolean"},defaultValue:{summary:"false"}}},className:{control:"text",description:"Additional CSS class names to apply to the `DialogContent` container."}},args:{DialogOverlayProps:{blurBackground:!1},maxWidth:"sm"},render:t=>e.jsxs(s,{children:[e.jsx(c,{asChild:!0,children:e.jsx(o,{variant:"primary",children:"Default Dialog"})}),e.jsx(d,{children:e.jsxs(l,{...t,children:[e.jsx(g,{children:"Are you absolutely sure?"}),e.jsx(j,{children:"Something that we want to hide visually but still want in the DOM for accessibility"}),e.jsx(u,{children:"This action cannot be undone. This will permanently delete your account and remove your data from our servers."}),e.jsxs(h,{className:"flex justify-end gap-4",children:[e.jsx(p,{asChild:!0,children:e.jsx(o,{variant:"secondary",children:"Cancel"})}),e.jsx(o,{children:"Submit"})]})]})})]})},a={parameters:{docs:{description:{story:"A standard confirmation dialog composed of a `DialogTitle`, a `DialogDescription`, and a `DialogActions` footer. Clicking the trigger opens the dialog. Open dialogs may be dismissed either by clicking the `DialogClose` button or by clicking outside the dialog on the overlay. Use `VisuallyHidden` to provide accessible content (such as an alternate description) that should not appear visually."},source:{code:`
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Default Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent>
            <DialogTitle>Are you absolutely sure?</DialogTitle>
            <VisuallyHidden>
                Something that we want to hide visually but still want in the DOM for accessibility
            </VisuallyHidden>
            <DialogDescription>
                This action cannot be undone. This will permanently delete your account and remove your data
                from our servers.
            </DialogDescription>
            <DialogActions className='flex justify-end gap-4'>
                <DialogClose asChild>
                    <Button variant='secondary'>Cancel</Button>
                </DialogClose>
                <Button>Submit</Button>
            </DialogActions>
        </DialogContent>
    </DialogPortal>
</Dialog>`,language:"tsx",type:"code"}}}},A="Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",n={parameters:{docs:{description:{story:"When the body content exceeds the available height, only the middle section scrolls — the `DialogTitle` stays pinned to the top of the dialog and the `DialogActions` footer stays pinned to the bottom. The dialog is constrained by `max-h-[calc(100vh-32px)]` and any children other than the title and actions are placed inside an internal scrollable container. This works for both `DialogDescription` and arbitrary JSX children."},source:{code:`
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Overflowing Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent>
            <DialogTitle>Terms and Conditions</DialogTitle>
            <DialogDescription>A bunch of repeating text</DialogDescription>
            {Array.from({ length: 25 }, (_, index) => (
                <p key={index} className='mb-4'>
                    {loremParagraph}
                </p>
            ))}
            <DialogActions className='flex justify-end gap-4'>
                <DialogClose asChild>
                    <Button variant='secondary'>Cancel</Button>
                </DialogClose>
                <Button>Accept</Button>
            </DialogActions>
        </DialogContent>
    </DialogPortal>
</Dialog>`,language:"tsx",type:"code"}}},render:t=>e.jsxs(s,{children:[e.jsx(c,{asChild:!0,children:e.jsx(o,{variant:"primary",children:"Overflowing Dialog"})}),e.jsx(d,{children:e.jsxs(l,{...t,children:[e.jsx(g,{children:"Terms and Conditions"}),e.jsx(u,{children:"A bunch of repeating text"}),Array.from({length:25},(B,r)=>e.jsx("p",{className:"mb-4",children:A},r)),e.jsxs(h,{className:"flex justify-end gap-4",children:[e.jsx(p,{asChild:!0,children:e.jsx(o,{variant:"secondary",children:"Cancel"})}),e.jsx(o,{children:"Accept"})]})]})})]})},i={parameters:{docs:{description:{story:"Setting `disableStickyLayout` to `true` opts out of the default sticky layout behavior. The `DialogTitle` and `DialogActions` are no longer pinned to the top and bottom of the dialog, and children are rendered in source order without an internal scrollable container. When the body content exceeds the available height the entire `DialogContent` scrolls as a single block, so the title and actions scroll out of view along with the rest of the content. Use this when you need full control over the dialog layout."},source:{code:`
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Overflowing Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent disableStickyLayout>
            <DialogTitle>Terms and Conditions</DialogTitle>
            <DialogDescription>A bunch of repeating text</DialogDescription>
            {Array.from({ length: 25 }, (_, index) => (
                <p key={index} className='mb-4'>
                    {loremParagraph}
                </p>
            ))}
            <DialogActions className='flex justify-end gap-4'>
                <DialogClose asChild>
                    <Button variant='secondary'>Cancel</Button>
                </DialogClose>
                <Button>Accept</Button>
            </DialogActions>
        </DialogContent>
    </DialogPortal>
</Dialog>`,language:"tsx",type:"code"}}},args:{disableStickyLayout:!0},render:t=>e.jsxs(s,{children:[e.jsx(c,{asChild:!0,children:e.jsx(o,{variant:"primary",children:"Overflowing Dialog without auto "})}),e.jsx(d,{children:e.jsxs(l,{...t,children:[e.jsx(g,{children:"Terms and Conditions"}),e.jsx(u,{children:"A bunch of repeating text"}),Array.from({length:25},(B,r)=>e.jsx("p",{className:"mb-4",children:A},r)),e.jsxs(h,{className:"flex justify-end gap-4",children:[e.jsx(p,{asChild:!0,children:e.jsx(o,{variant:"secondary",children:"Cancel"})}),e.jsx(o,{children:"Accept"})]})]})})]})};var D,m,y;a.parameters={...a.parameters,docs:{...(D=a.parameters)==null?void 0:D.docs,source:{originalSource:`{
  parameters: {
    docs: {
      description: {
        story: 'A standard confirmation dialog composed of a \`DialogTitle\`, a \`DialogDescription\`, and a \`DialogActions\` footer. Clicking the trigger opens the dialog. Open dialogs may be dismissed either by clicking the \`DialogClose\` button or by clicking outside the dialog on the overlay. Use \`VisuallyHidden\` to provide accessible content (such as an alternate description) that should not appear visually.'
      },
      source: {
        code: \`
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Default Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent>
            <DialogTitle>Are you absolutely sure?</DialogTitle>
            <VisuallyHidden>
                Something that we want to hide visually but still want in the DOM for accessibility
            </VisuallyHidden>
            <DialogDescription>
                This action cannot be undone. This will permanently delete your account and remove your data
                from our servers.
            </DialogDescription>
            <DialogActions className='flex justify-end gap-4'>
                <DialogClose asChild>
                    <Button variant='secondary'>Cancel</Button>
                </DialogClose>
                <Button>Submit</Button>
            </DialogActions>
        </DialogContent>
    </DialogPortal>
</Dialog>\`,
        language: 'tsx',
        type: 'code'
      }
    }
  }
}`,...(y=(m=a.parameters)==null?void 0:m.docs)==null?void 0:y.source}}};var f,b,x;n.parameters={...n.parameters,docs:{...(f=n.parameters)==null?void 0:f.docs,source:{originalSource:`{
  parameters: {
    docs: {
      description: {
        story: 'When the body content exceeds the available height, only the middle section scrolls — the \`DialogTitle\` stays pinned to the top of the dialog and the \`DialogActions\` footer stays pinned to the bottom. The dialog is constrained by \`max-h-[calc(100vh-32px)]\` and any children other than the title and actions are placed inside an internal scrollable container. This works for both \`DialogDescription\` and arbitrary JSX children.'
      },
      source: {
        code: \`
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Overflowing Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent>
            <DialogTitle>Terms and Conditions</DialogTitle>
            <DialogDescription>A bunch of repeating text</DialogDescription>
            {Array.from({ length: 25 }, (_, index) => (
                <p key={index} className='mb-4'>
                    {loremParagraph}
                </p>
            ))}
            <DialogActions className='flex justify-end gap-4'>
                <DialogClose asChild>
                    <Button variant='secondary'>Cancel</Button>
                </DialogClose>
                <Button>Accept</Button>
            </DialogActions>
        </DialogContent>
    </DialogPortal>
</Dialog>\`,
        language: 'tsx',
        type: 'code'
      }
    }
  },
  render: args => {
    return <Dialog>
                <DialogTrigger asChild>
                    <Button variant='primary'>Overflowing Dialog</Button>
                </DialogTrigger>
                <DialogPortal>
                    <DialogContent {...args}>
                        <DialogTitle>Terms and Conditions</DialogTitle>
                        <DialogDescription>A bunch of repeating text</DialogDescription>
                        {Array.from({
            length: 25
          }, (_, index) => <p key={index} className='mb-4'>
                                {loremParagraph}
                            </p>)}
                        <DialogActions className='flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button variant='secondary'>Cancel</Button>
                            </DialogClose>
                            <Button>Accept</Button>
                        </DialogActions>
                    </DialogContent>
                </DialogPortal>
            </Dialog>;
  }
}`,...(x=(b=n.parameters)==null?void 0:b.docs)==null?void 0:x.source}}};var C,v,T;i.parameters={...i.parameters,docs:{...(C=i.parameters)==null?void 0:C.docs,source:{originalSource:`{
  parameters: {
    docs: {
      description: {
        story: 'Setting \`disableStickyLayout\` to \`true\` opts out of the default sticky layout behavior. The \`DialogTitle\` and \`DialogActions\` are no longer pinned to the top and bottom of the dialog, and children are rendered in source order without an internal scrollable container. When the body content exceeds the available height the entire \`DialogContent\` scrolls as a single block, so the title and actions scroll out of view along with the rest of the content. Use this when you need full control over the dialog layout.'
      },
      source: {
        code: \`
<Dialog>
    <DialogTrigger asChild>
        <Button variant='primary'>Overflowing Dialog</Button>
    </DialogTrigger>
    <DialogPortal>
        <DialogContent disableStickyLayout>
            <DialogTitle>Terms and Conditions</DialogTitle>
            <DialogDescription>A bunch of repeating text</DialogDescription>
            {Array.from({ length: 25 }, (_, index) => (
                <p key={index} className='mb-4'>
                    {loremParagraph}
                </p>
            ))}
            <DialogActions className='flex justify-end gap-4'>
                <DialogClose asChild>
                    <Button variant='secondary'>Cancel</Button>
                </DialogClose>
                <Button>Accept</Button>
            </DialogActions>
        </DialogContent>
    </DialogPortal>
</Dialog>\`,
        language: 'tsx',
        type: 'code'
      }
    }
  },
  args: {
    disableStickyLayout: true
  },
  render: args => {
    return <Dialog>
                <DialogTrigger asChild>
                    <Button variant='primary'>Overflowing Dialog without auto </Button>
                </DialogTrigger>
                <DialogPortal>
                    <DialogContent {...args}>
                        <DialogTitle>Terms and Conditions</DialogTitle>
                        <DialogDescription>A bunch of repeating text</DialogDescription>
                        {Array.from({
            length: 25
          }, (_, index) => <p key={index} className='mb-4'>
                                {loremParagraph}
                            </p>)}
                        <DialogActions className='flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button variant='secondary'>Cancel</Button>
                            </DialogClose>
                            <Button>Accept</Button>
                        </DialogActions>
                    </DialogContent>
                </DialogPortal>
            </Dialog>;
  }
}`,...(T=(v=i.parameters)==null?void 0:v.docs)==null?void 0:T.source}}};const Z=["DefaultDialog","OverflowingContent","OverflowingContentWithDisabledStickyLayout"];export{a as DefaultDialog,n as OverflowingContent,i as OverflowingContentWithDisabledStickyLayout,Z as __namedExportsOrder,Y as default};
