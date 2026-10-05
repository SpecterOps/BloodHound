import{j as o}from"./jsx-runtime-BTJTZTIL.js";import{F as n,f as N,a as W,b as D,c as U,d as O}from"./index.es-C0EmWI2V.js";import{B as r}from"./Badge-ldcobJiH.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-2AqrBsEA.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";const V={chevronUp:o.jsx(n,{icon:O}),chevronDown:o.jsx(n,{icon:U}),plus:o.jsx(n,{icon:D}),minus:o.jsx(n,{icon:W}),eyeSlash:o.jsx(n,{icon:N})},$={chevronUp:"<FontAwesomeIcon icon={faChevronUp} />",chevronDown:"<FontAwesomeIcon icon={faChevronDown} />",plus:"<FontAwesomeIcon icon={faPlus} />",minus:"<FontAwesomeIcon icon={faMinus} />",eyeSlash:"<FontAwesomeIcon icon={faEyeSlash} />"},C=(e,d)=>{const{iconName:m,...a}=d.args,p=m?$[m]:void 0;return`<Badge ${[`label="${a.label}"`,a.variant&&a.variant!=="fill"&&`variant="${a.variant}"`,a.color&&`color="${a.color}"`,p&&a.iconPosition&&`iconPosition="${a.iconPosition}"`,p&&`icon={${p}}`].filter(Boolean).join(" ")} />`},u={iconName:{control:!1,table:{disable:!0}},iconPosition:{control:!1,table:{disable:!0}},iconClassName:{control:!1,table:{disable:!0}}},J={title:"Components/Badge",component:r,tags:["autodocs"],parameters:{layout:"centered",docs:{description:{component:`
Badges are used to communicate **status, counts, or small pieces of metadata** in a compact, visually distinct way.

### When to use
- To display **status** (e.g., active, inactive, warning)
- To show **counts or indicators** (e.g., notifications, totals)
- To highlight **categorization or labels** (e.g., tags, types)

Badges are best used sparingly and should remain **short, scannable, and non-interactive**.
                `},canvas:{sourceState:"shown"}}},argTypes:{variant:{control:"select",options:["fill","outline"],description:"Visual style of the badge.",table:{type:{summary:"'fill' | 'outline'"},defaultValue:{summary:"outline"}}},label:{control:"text",description:"Text displayed inside the badge.",table:{type:{summary:"string"}}},color:{control:"select",options:["primary","secondary","grey","red","orange","green","blue","error","warning","success","disabled"],description:"Color of the badge. Hex colors are deprecated — use a named color instead.",table:{type:{summary:"'primary' | 'secondary' | 'grey' | 'red' | 'orange' | 'green' | 'blue'"},defaultValue:{summary:"grey"}}},iconPosition:{control:"select",options:["left","right"],description:"Position of the icon relative to the label. Only applies when an icon is provided.",table:{type:{summary:"'left' | 'right'"},defaultValue:{summary:"left"}}},iconName:{name:"icon",control:"select",options:["chevronUp","chevronDown","plus","minus","eyeSlash"],labels:{chevronUp:"Chevron Up",chevronDown:"Chevron Down",plus:"Plus",minus:"Minus",eyeSlash:"Eye Slash"},description:"Select an icon to display.",table:{type:{summary:"ReactNode"},defaultValue:{summary:"undefined"}}},icon:{control:!1,table:{disable:!0}},iconClassName:{control:"text",description:"Additional class names applied to the icon wrapper.",table:{type:{summary:"string"},defaultValue:{summary:"undefined"}}}},args:{label:"Badge",variant:"outline",color:"grey",iconPosition:"left"}},F=({iconName:e,...d})=>o.jsx(r,{...d,icon:e?V[e]:void 0}),s={tags:["!dev"],render:F,parameters:{docs:{source:{transform:C}}}},t={args:{label:"No Icon",iconName:void 0},argTypes:u,render:e=>o.jsx(r,{...e})},i={args:{label:"With Icon",iconName:"chevronUp",iconPosition:"left"},render:F,parameters:{docs:{source:{transform:C}}}},c={args:{label:"Fill",variant:"fill",color:"primary"},argTypes:u,render:e=>o.jsx(r,{...e})},l={args:{label:"Outline",variant:"outline",color:"primary"},argTypes:u,render:e=>o.jsx(r,{...e})};var g,y,f;s.parameters={...s.parameters,docs:{...(g=s.parameters)==null?void 0:g.docs,source:{originalSource:`{
  tags: ['!dev'],
  render: renderWithIcon,
  parameters: {
    docs: {
      source: {
        transform: buildBadgeSource
      }
    }
  }
}`,...(f=(y=s.parameters)==null?void 0:y.docs)==null?void 0:f.source}}};var h,b,v;t.parameters={...t.parameters,docs:{...(h=t.parameters)==null?void 0:h.docs,source:{originalSource:`{
  args: {
    label: 'No Icon',
    iconName: undefined
  },
  argTypes: disabledIconArgTypes,
  render: args => <Badge {...args} />
}`,...(v=(b=t.parameters)==null?void 0:b.docs)==null?void 0:v.source}}};var I,w,S;i.parameters={...i.parameters,docs:{...(I=i.parameters)==null?void 0:I.docs,source:{originalSource:`{
  args: {
    label: 'With Icon',
    iconName: 'chevronUp',
    iconPosition: 'left'
  },
  render: renderWithIcon,
  parameters: {
    docs: {
      source: {
        transform: buildBadgeSource
      }
    }
  }
}`,...(S=(w=i.parameters)==null?void 0:w.docs)==null?void 0:S.source}}};var x,T,B;c.parameters={...c.parameters,docs:{...(x=c.parameters)==null?void 0:x.docs,source:{originalSource:`{
  args: {
    label: 'Fill',
    variant: 'fill',
    color: 'primary'
  },
  argTypes: disabledIconArgTypes,
  render: args => <Badge {...args} />
}`,...(B=(T=c.parameters)==null?void 0:T.docs)==null?void 0:B.source}}};var j,P,A;l.parameters={...l.parameters,docs:{...(j=l.parameters)==null?void 0:j.docs,source:{originalSource:`{
  args: {
    label: 'Outline',
    variant: 'outline',
    color: 'primary'
  },
  argTypes: disabledIconArgTypes,
  render: args => <Badge {...args} />
}`,...(A=(P=l.parameters)==null?void 0:P.docs)==null?void 0:A.source}}};const K=["Default","WithoutIcon","WithIcon","Fill","Outline"];export{s as Default,c as Fill,l as Outline,i as WithIcon,t as WithoutIcon,K as __namedExportsOrder,J as default};
