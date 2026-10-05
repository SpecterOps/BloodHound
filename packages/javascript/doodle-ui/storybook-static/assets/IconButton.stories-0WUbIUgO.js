import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{F as I,j as v}from"./index.es-C0EmWI2V.js";import{A as S}from"./AppIcons-Dp2sojKC.js";import{A as B}from"./AccessibleIcon-BlGV1aYS.js";import{I as i}from"./IconButton-CiUmUVXd.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-2AqrBsEA.js";import"./Card-BdBAz3fL.js";import"./Typography-BMbK6Q6M.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";import"./UserCog-VADywgBc.js";import"./VisuallyHidden-DT_8ZbEK.js";import"./index-oIeaYTeP.js";import"./index-Dc_EP5I0.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./Tooltip-4xw-YYUR.js";import"./index-Dv0R_RYs.js";import"./index-BWEBPL6D.js";import"./index-DnVAETpg.js";import"./index-ejpvWIAw.js";import"./index-DSMk1t8m.js";import"./floating-ui.react-dom-D40P0bPq.js";import"./index-DRLFypRQ.js";import"./index-vQUXxau3.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";const te={title:"Components/IconButton",component:i,tags:["autodocs"],args:{variant:"default",disabled:!1,hideTooltip:!1,size:16,"aria-label":"More options",children:e.jsx(I,{icon:v})},argTypes:{variant:{options:["default","primary","secondary"],control:"select",description:"Sets the visual emphasis of the button.",table:{category:"Appearance",defaultValue:{summary:"default"}}},children:{control:!1,description:"A single AppIcon, Font Awesome icon, or AccessibleIcon element.",table:{category:"Content"}},"aria-label":{description:"Required accessible name describing the action performed by the icon button.",control:"text",table:{category:"Accessibility",type:{summary:"string"}}},tooltip:{description:"Visible tooltip content. Defaults to aria-label; override it when the tooltip needs additional context.",control:"text",table:{category:"Accessibility"}},hideTooltip:{description:"Suppresses the tooltip without removing the button accessible label.",control:"boolean",table:{category:"Accessibility",defaultValue:{summary:"false"}}},size:{description:"Sets the icon width and height in pixels. Defaults to 16. The square button resizes with the icon.",control:{type:"number",min:8,step:1},table:{category:"Appearance",defaultValue:{summary:"16"},type:{summary:"number"}}},disabled:{control:"boolean",description:"Disables button interaction. The tooltip remains available unless hideTooltip is true.",table:{category:"State",defaultValue:{summary:"false"}}},className:{control:"text",description:"Classes applied to the button element.",table:{category:"Appearance"}}},parameters:{layout:"centered",docs:{description:{component:"Use IconButton for an action represented only by an icon. If the action also has visible text, use Button or TextButton instead.\n\n### Basic usage\n\nProvide one icon child and an `aria-label` that describes the action. The button uses the label as its accessible name and default tooltip. Describe what happens when the button is activated—not the icon's appearance.\n\n```tsx\n<IconButton aria-label='Show filter options'>\n    <AppIcon.FilterOutline />\n</IconButton>\n```\n\nPrefer `\"Show filter options\"` over `\"Filter icon\"`. IconButton marks its child as decorative because the button itself owns the accessible name.\n\n### Supported icons\n\nIconButton accepts one `AppIcon`, Font Awesome icon, or `AccessibleIcon` element. Passing an icon directly is the simplest option. If an `AccessibleIcon` is nested, IconButton suppresses its label and tooltip to prevent duplicate announcements and tooltips.\n\n```tsx\n<IconButton aria-label='More options'>\n    <FontAwesomeIcon icon={faEllipsisVertical} />\n</IconButton>\n\n<IconButton aria-label='Show information'>\n    <AccessibleIcon label='Information'>\n        <AppIcon.Info />\n    </AccessibleIcon>\n</IconButton>\n```\n\n### Tooltip behavior\n\nThe tooltip defaults to `aria-label`. Use `tooltip` when the visible message needs extra context while the concise accessible name should remain unchanged. Use `hideTooltip` only when the action is already identified by nearby visible text.\n\n```tsx\n<IconButton\n    aria-label='Show filter options'\n    tooltip='Filters are unavailable while data loads'\n    disabled\n>\n    <AppIcon.FilterOutline />\n</IconButton>\n```\n\nDisabled buttons retain their tooltip so users can still learn what the action does or why it is unavailable. Set `hideTooltip` to suppress it explicitly.\n\n### Variants\n\n- `default` is transparent and works well for standalone controls and toolbars.\n- `primary` gives the action the strongest emphasis.\n- `secondary` provides emphasis without competing with a primary action.\n\nUse the `variant` prop rather than applying Button variant classes through `className`.\n\n### Sizing and color\n\nThe `size` prop controls the icon's width and height in pixels. It defaults to `16`. IconButton adds `8px` of padding on each side, so `size={16}` produces a `32px × 32px` button.\n\nIcons inherit the button's text color through `currentColor`. Prefer a variant for standard colors and use `className` only for a deliberate local override.\n\n```tsx\n<IconButton aria-label='More options' size={20} className='text-primary'>\n    <FontAwesomeIcon icon={faEllipsisVertical} />\n</IconButton>\n```"}}}},r=s=>({children:w,...A})=>e.jsxs(e.Fragment,{children:[e.jsx("div",{className:"flex justify-center mb-10",children:e.jsx(i,{...A,children:w})}),e.jsx("hr",{className:"mb-10"}),e.jsxs("div",{className:"flex items-center justify-center gap-8",children:[e.jsxs("div",{className:"flex flex-col items-center gap-4",children:[e.jsx(i,{"aria-label":"More options",size:18,variant:s,children:e.jsx(B,{label:"More options",children:e.jsx(I,{icon:v})})}),"Enabled"]}),e.jsxs("div",{className:"flex flex-col items-center gap-4",children:[e.jsx(i,{"aria-label":"Show filter options",disabled:!0,size:18,variant:s,children:e.jsx(S.FilterOutline,{})}),"Disabled"]})]})]}),t={parameters:{docs:{description:{story:"Use the transparent default variant for icon actions in toolbars or other compact control groups."}}},render:r("default")},o={args:{variant:"primary"},parameters:{docs:{description:{story:"Use the primary variant when the icon action is the main action in its immediate context."}}},render:r("primary")},a={args:{variant:"secondary"},parameters:{docs:{description:{story:"Use the secondary variant for a supporting icon action alongside a primary action."}}},render:r("secondary")},n={args:{hideTooltip:!0},parameters:{docs:{description:{story:"Hide the tooltip only when nearby visible content already identifies the action. The button keeps its accessible name."}}}};var c,l,p;t.parameters={...t.parameters,docs:{...(c=t.parameters)==null?void 0:c.docs,source:{originalSource:`{
  parameters: {
    docs: {
      description: {
        story: 'Use the transparent default variant for icon actions in toolbars or other compact control groups.'
      }
    }
  },
  render: renderVariantStory('default')
}`,...(p=(l=t.parameters)==null?void 0:l.docs)==null?void 0:p.source}}};var d,m,u;o.parameters={...o.parameters,docs:{...(d=o.parameters)==null?void 0:d.docs,source:{originalSource:`{
  args: {
    variant: 'primary'
  },
  parameters: {
    docs: {
      description: {
        story: 'Use the primary variant when the icon action is the main action in its immediate context.'
      }
    }
  },
  render: renderVariantStory('primary')
}`,...(u=(m=o.parameters)==null?void 0:m.docs)==null?void 0:u.source}}};var h,b,y;a.parameters={...a.parameters,docs:{...(h=a.parameters)==null?void 0:h.docs,source:{originalSource:`{
  args: {
    variant: 'secondary'
  },
  parameters: {
    docs: {
      description: {
        story: 'Use the secondary variant for a supporting icon action alongside a primary action.'
      }
    }
  },
  render: renderVariantStory('secondary')
}`,...(y=(b=a.parameters)==null?void 0:b.docs)==null?void 0:y.source}}};var f,g,x;n.parameters={...n.parameters,docs:{...(f=n.parameters)==null?void 0:f.docs,source:{originalSource:`{
  args: {
    hideTooltip: true
  },
  parameters: {
    docs: {
      description: {
        story: 'Hide the tooltip only when nearby visible content already identifies the action. The button keeps its accessible name.'
      }
    }
  }
}`,...(x=(g=n.parameters)==null?void 0:g.docs)==null?void 0:x.source}}};const oe=["Default","Primary","Secondary","TooltipHidden"];export{t as Default,o as Primary,a as Secondary,n as TooltipHidden,oe as __namedExportsOrder,te as default};
