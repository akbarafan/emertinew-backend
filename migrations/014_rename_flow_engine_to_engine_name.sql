-- Engine label for every sensor of an MCU, not only the flow meters.
--
-- The flow meters were already labelled with the engine they belong to, which
-- is what groups them on the command center. The tachometers had no such label,
-- and the channel index cannot stand in for one: the flow channels are ordered
-- by the flow meters of the tanks (the .env of a vessel names them "A Biru",
-- "B Biru", ...) while the RPM channels are ordered by the tachometer pins
-- (RPM_GPIO=23,24 against FLOW_GPIO=5,6). The two orders are independent, so
-- matching them by position was a guess.
--
-- Renaming the column makes the label apply to any sensor, so an operator can
-- label the tachometers with the same engine name and the engine cards can
-- show a real RPM instead of a guess. A sensor that carries no label simply
-- has none, and the panel says so.
--
-- The existing labels are preserved by the rename.

ALTER TABLE m_sensors
    RENAME COLUMN flow_engine TO engine_name;

COMMENT ON COLUMN m_sensors.engine_name IS
    'Engine this sensor belongs to. Applies to the flow meters and the tachometers; the label is what groups the sensors on the command center.';
