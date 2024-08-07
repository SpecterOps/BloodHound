match (n) detach delete n;

create (n1:NodeKind1 {name: 'n1'})
create (n2:NodeKind1 {name: 'n2'}) set n2:NodeKind2
create (n3:NodeKind1 {name: 'n3'}) set n3:NodeKind2
create (n4:NodeKind2 {name: 'n4'})
create (n5:NodeKind2 {name: 'n5'})
create (n1)-[:EdgeKind1 {name: 'e1', prop: 'a'}]->(n2)-[:EdgeKind1 {name: 'e2', prop: 'a'}]->(n3)-[:EdgeKind1 {name: 'e3'}]->(n4);
