import{w as E,e as t,u as a}from"./index-Ch1O_M5-.js";import{S as D}from"./Switch-B-7ih0xl.js";import"./jsx-runtime-BTJTZTIL.js";import"./index-ChsGqxH_.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-Dv0R_RYs.js";import"./index-Dc_EP5I0.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-CFknkR0k.js";import"./index-DRLFypRQ.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";const K={title:"Components/Switch",component:D,tags:["autodocs"],argTypes:{checked:{control:"boolean",description:"The controlled checked state of the switch.",table:{type:{summary:"boolean"},required:!0}},onCheckedChange:{action:"checked changed",description:"Called when the checked state changes.",table:{type:{summary:"(checked: boolean) => void"},required:!0}},label:{control:"text"},labelPosition:{options:["left","right"],control:"select"}},args:{checked:!1,label:"Label",labelPosition:"right"},parameters:{layout:"centered",docs:{description:{component:`
The Switch component represents a binary on/off control.

By default, the switch is rendered in an unchecked (off) state and can be toggled by user interaction (click or keyboard). When a label is provided, it is associated to the control for improved accessibility and can be positioned to the left or right.

When disabled, the switch is non-interactive and cannot be toggled. Visual styles indicate the disabled state for both the switch and its label.

Use this component for immediate state changes, such as enabling features or toggling settings.
        `}}}},s={args:{checked:!1},play:async({canvasElement:S,step:l})=>{const e=await E(S).findByRole("switch");await l("Click the switch to toggle it",async()=>{t(e).toHaveAttribute("aria-checked","false"),await a.click(e),await t(e).toHaveAttribute("aria-checked","true")}),await l("Pressing spacebar or enter on the focused element toggles the switch",async()=>{t(e).toHaveAttribute("aria-checked","true"),e.focus(),await a.keyboard(" "),await t(e).toHaveAttribute("aria-checked","false"),e.focus(),await a.keyboard("{enter}"),await t(e).toHaveAttribute("aria-checked","true"),await a.keyboard("{tab}"),await t(e).not.toHaveAttribute("aria-checked","false")})}},r={args:{checked:!0}},c={args:{label:"Label"}},o={args:{label:"Label",labelPosition:"left"}},n={args:{disabled:!0,checked:!0,label:"Label"}},i={args:{disabled:!0,checked:!1,label:"Label"}};var d,h,u;s.parameters={...s.parameters,docs:{...(d=s.parameters)==null?void 0:d.docs,source:{originalSource:`{
  args: {
    checked: false
  },
  play: async ({
    canvasElement,
    step
  }) => {
    const canvas = within(canvasElement);
    const switchButton = await canvas.findByRole('switch');
    await step('Click the switch to toggle it', async () => {
      expect(switchButton).toHaveAttribute('aria-checked', 'false');
      await userEvent.click(switchButton);
      await expect(switchButton).toHaveAttribute('aria-checked', 'true');
    });
    await step('Pressing spacebar or enter on the focused element toggles the switch', async () => {
      expect(switchButton).toHaveAttribute('aria-checked', 'true');
      switchButton.focus();
      await userEvent.keyboard(' ');
      await expect(switchButton).toHaveAttribute('aria-checked', 'false');
      switchButton.focus();
      await userEvent.keyboard('{enter}');
      await expect(switchButton).toHaveAttribute('aria-checked', 'true');

      // Other keyboard inputs like tab should not toggle the switch
      await userEvent.keyboard('{tab}');
      await expect(switchButton).not.toHaveAttribute('aria-checked', 'false');
    });
  }
}`,...(u=(h=s.parameters)==null?void 0:h.docs)==null?void 0:u.source}}};var b,p,m;r.parameters={...r.parameters,docs:{...(b=r.parameters)==null?void 0:b.docs,source:{originalSource:`{
  args: {
    checked: true
  }
}`,...(m=(p=r.parameters)==null?void 0:p.docs)==null?void 0:m.source}}};var w,g,k;c.parameters={...c.parameters,docs:{...(w=c.parameters)==null?void 0:w.docs,source:{originalSource:`{
  args: {
    label: 'Label'
  }
}`,...(k=(g=c.parameters)==null?void 0:g.docs)==null?void 0:k.source}}};var f,y,v;o.parameters={...o.parameters,docs:{...(f=o.parameters)==null?void 0:f.docs,source:{originalSource:`{
  args: {
    label: 'Label',
    labelPosition: 'left'
  }
}`,...(v=(y=o.parameters)==null?void 0:y.docs)==null?void 0:v.source}}};var L,B,A;n.parameters={...n.parameters,docs:{...(L=n.parameters)==null?void 0:L.docs,source:{originalSource:`{
  args: {
    disabled: true,
    checked: true,
    label: 'Label'
  }
}`,...(A=(B=n.parameters)==null?void 0:B.docs)==null?void 0:A.source}}};var H,x,C;i.parameters={...i.parameters,docs:{...(H=i.parameters)==null?void 0:H.docs,source:{originalSource:`{
  args: {
    disabled: true,
    checked: false,
    label: 'Label'
  }
}`,...(C=(x=i.parameters)==null?void 0:x.docs)==null?void 0:C.source}}};const M=["Default","Checked","Labeled","LeftLabeled","DisabledChecked","DisabledUnchecked"];export{r as Checked,s as Default,n as DisabledChecked,i as DisabledUnchecked,c as Labeled,o as LeftLabeled,M as __namedExportsOrder,K as default};
