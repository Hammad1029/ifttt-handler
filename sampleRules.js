const x = {
  apiName: "myApiiii",
  apiPath: "/createC",
  apiGroup: "group1",
  apiDescription: "SDSDA",
  startRules: [0],
  rules: [
    {
      op1: {
        type: "req",
        data: {
          get: "a.b.c",
        },
      },
      opnd: "eq",
      op2: {
        type: "db",
        data: {
          get: "first",
          query: {
            type: "SELECT",
            table: "table_table1_010d6515_0ade_11ef_aaaa_001",
            conditions: {
              int1: {
                eq: {
                  type: "req",
                  data: {
                    get: "a.b.c",
                  },
                },
              },
              text2: "yooo",
            },
            projections: {
              int1: {
                as: "first",
                mutate: [],
              },
            },
            columns: {},
          },
        },
      },
      then: [
        {
          type: "rule",
          data: {
            value: 1,
          },
        },
        {
          type: "setRes",
          data: {
            res: "00",
          },
        },
      ],
      else: [
        {
          type: "setRes",
          data: {
            res: "500",
          },
        },
        {
          type: "db",
          data: {
            type: "INSERT",
            table: "table_table1_010d6515_0ade_11ef_aaaa_001",
            columns: {
              int1: 1,
              text2: "yooo",
              int2: 3,
              text1: "hello",
              int13: 2,
            },
            conditions: {},
            projections: {},
          },
        },
        {
          type: "res",
          data: {},
        },
      ],
    },
    {
      op1: {
        type: "const",
        data: {
          get: "true",
        },
      },
      opnd: "eq",
      op2: {
        type: "const",
        data: {
          get: "true",
        },
      },
      then: [
        {
          type: "setRes",
          data: {
            res2: "00",
          },
        },
        {
          type: "db",
          data: {
            query: {
              type: "DELETE",
              table: "table_table1_010d6515_0ade_11ef_aaaa_001",
              columns: {
                int1: {
                  eq: {
                    type: "req",
                    data: {
                      get: "a.b.c",
                    },
                  },
                },
                text2: "yooo",
              },
              conditions: {},
              projections: {},
            },
          },
        },
        {
          type: "res",
          data: {},
        },
      ],
      else: [],
    },
  ],
};
