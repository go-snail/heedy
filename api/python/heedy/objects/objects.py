from typing import Dict

from ..base import APIObject, APIList, Session
from ..kv import KV

from .. import users
from .. import apps
from ..notifications import Notifications

from . import registry


class Object(APIObject):
    props = {"name", "description", "icon", "meta", "tags", "key"}

    def __init__(self, objectData: Dict, session: Session):
        super().__init__(
            f"api/objects/{objectData['id']}",
            {"object": objectData["id"]},
            session,
            cached_data=objectData
        )
        self._kv = KV(f"api/kv/objects/{objectData['id']}", self.session)

    @property
    def kv(self):
        return self._kv

    @kv.setter
    def kv(self, v):
        return self._kv.set(**v)

    def __getattr__(self, attr):
        try:
            return self.cached_data[attr]
        except:
            return None

    @property
    def owner(self):
        return users.User(self.cached_data["owner"], self.session)

    @property
    def app(self):
        if self.cached_data["app"] is None:
            return None
        return apps.App(self.cached_data["app"], session=self.session)


class Objects(APIList):
    def __init__(self, constraints: Dict, session: Session):
        super().__init__("api/objects", constraints, session)

    def __getitem__(self, item):
        return super()._getitem(item, f=lambda x: registry.getObject(x, self.session))

    def __call__(self, **kwargs):
        # To query by object type, we use otype
        if "otype" in kwargs:
            kwargs["type"] = kwargs["otype"]
            del kwargs["otype"]
        return super()._call(
            f=lambda x: [registry.getObject(xx, self.session) for xx in x], **kwargs
        )

    def create(self, name, meta={}, otype="timeseries", **kwargs):
        """
        Creates a new object of the given type (timeseries by default).
        """
        return super()._create(
            f=lambda x: registry.getObject(x, self.session),
            **{"name": name, "type": otype, "meta": meta, **kwargs},
        )
